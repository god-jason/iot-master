package gateway

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/god-jason/iot-master/internal/iot"
	"github.com/god-jason/iot-master/pkg/logger"
	"github.com/god-jason/iot-master/pkg/thing"
)

// pollJob 一条链路在本轮需要采集的设备集合
type pollJob struct {
	link    *Link     // 目标链路（同一链路的任务串行执行）
	targets []*Device // 本轮待采集的设备快照
}

// Poller 每个 Server 一个轮询器
//
// 设计：
//   - ticker 周期遍历 Server 的 links
//   - 到期链路打包成一个 pollJob，交由 workers（=GatewayServer.Concurrency）并发消费
//   - worker 调用该链路初始创建的 protocol.Protocol.Poll(deviceId)
//   - 同一链路的任务串行（Link.busy），不同链路之间并发
type Poller struct {
	srv             *Server       // 所属 Server，用于遍历其 links
	workers         int           // 并发消费任务的 worker 数量
	tick            time.Duration // 调度轮询周期（遍历 links 的间隔）
	defaultInterval time.Duration // 链路未单独配置时的默认采集间隔

	jobCh    chan pollJob   // 待执行任务队列（有界，满则跳过本轮）
	done     chan struct{}  // 停止信号
	stopOnce sync.Once      // 保证 Stop 只关闭一次 done
	started  bool           // Run 是否已进入启动流程（mu 保护）
	mu       sync.Mutex     // 保护 started
	wg       sync.WaitGroup // 等待所有 worker 退出
	dropped  uint64         // 因队列满而丢弃的轮次计数（原子访问，用于观测调度积压）
}

// NewPoller 创建轮询器（不启动，显式 Run）
func NewPoller(srv *Server, workers int, tick, defaultInterval time.Duration) *Poller {
	// 兜底正常化参数，避免 0 值导致 ticker/worker 不可用
	if workers <= 0 {
		workers = 1
	}
	if tick <= 0 {
		tick = time.Second
	}
	if defaultInterval <= 0 {
		defaultInterval = 5 * time.Second
	}
	return &Poller{
		srv:             srv,
		workers:         workers,
		tick:            tick,
		defaultInterval: defaultInterval,
		// 队列容量按 worker 数的 4 倍设计，允许短时突发而不阻塞调度
		jobCh: make(chan pollJob, workers*4),
		done:  make(chan struct{}),
	}
}

// Run 启动轮询器（阻塞，应 go Poller.Run()）
func (p *Poller) Run() {
	// 幂等保护：重复调用 Run 直接返回
	p.mu.Lock()
	if p.started {
		p.mu.Unlock()
		return
	}
	p.started = true
	p.mu.Unlock()

	// 固定数量的 worker 消费任务队列
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for job := range p.jobCh {
				p.pollOnce(job)
			}
		}()
	}

	ticker := time.NewTicker(p.tick)
	defer ticker.Stop()

	for {
		select {
		case <-p.done:
			// 先关队列让 worker 排空退出，再等全部结束
			close(p.jobCh)
			p.wg.Wait()
			return
		case <-ticker.C:
			p.tickRound()
		}
	}
}

// Stop 停止轮询器
func (p *Poller) Stop() {
	p.stopOnce.Do(func() { close(p.done) })
}

// tickRound 遍历所有链路，派发到期的轮询任务
func (p *Poller) tickRound() {
	now := time.Now()

	for _, link := range p.srv.Links() {
		link.mu.Lock()
		// 协议实例未初始化（连接尚未 Attach）或已随连接关闭摘除：不参与轮询
		// 未注册、无设备或上一轮任务仍在执行：跳过本轮（保证同链路串行）
		if link.Proto == nil || !link.Registered || len(link.devices) == 0 || link.busy {
			link.mu.Unlock()
			continue
		}

		// 链路级配置优先，否则使用默认间隔
		interval := p.defaultInterval
		if link.pollCfg != nil && link.pollCfg.Interval > 0 {
			interval = link.pollCfg.Interval
		}
		// 未到采集时间，等待下一 tick
		if now.Sub(link.lastPoll) < interval {
			link.mu.Unlock()
			continue
		}

		link.lastPoll = now
		// 拷贝设备快照后在锁外执行任务，避免长时间持锁
		targets := append([]*Device(nil), link.devices...)
		link.busy = true
		link.mu.Unlock()

		// 队列满时本次任务直接作废，立即释放 busy 以便下一 tick 重试
		if !p.dispatch(pollJob{link: link, targets: targets}) {
			p.finish(link)
		}
	}
}

// dispatch 非阻塞入队：队列满时返回 false，由调用方放弃本轮任务
func (p *Poller) dispatch(job pollJob) bool {
	select {
	case p.jobCh <- job:
		return true
	default:
		// 队列满时丢弃本轮任务：计数并告警，便于发现 worker 不足或链路过多
		dropped := atomic.AddUint64(&p.dropped, 1)
		logger.Warn("软网关轮询队列已满，丢弃本轮任务", "server_id", p.srv.ServerID(), "link_id", job.link.ID(), "dropped", dropped)
		return false
	}
}

// finish 标记连接本轮任务完成
func (p *Poller) finish(link *Link) {
	link.mu.Lock()
	link.busy = false
	link.mu.Unlock()
}

// pollOnce worker 消费：对该链路的每个设备调用协议实例 Poll
func (p *Poller) pollOnce(job pollJob) {
	// 无论任务成功与否都要释放 busy，否则该链路将永久停止调度
	defer p.finish(job.link)

	if job.link == nil {
		return
	}

	// 逐设备串行采集：协议实例内部通常会做请求-响应关联
	for _, target := range job.targets {
		// 跳过已被卸载的设备（设备列表可能在任务排队期间被热更新）
		if target == nil || !target.Mounted() {
			continue
		}
		// 每个设备取一次协议实例：连接可能已在任务排队/执行期间关闭并摘除协议实例
		proto := job.link.ProtocolInstance()
		if proto == nil {
			return
		}
		values, err := proto.Poll(target.DeviceID(), pollTimeout)
		if err != nil {
			logger.Warn("软网关轮询失败", "server_id", job.link.ServerID, "link_id", job.link.ID(), "device_id", target.DeviceID(), "error", err)
			continue
		}
		if len(values) == 0 {
			continue
		}

		// 软网关后处理：表达式计算 + Filter 拦截
		if m, ok := iot.GetModel(target.ModelID()); ok {
			// PostProcess 返回 false 表示被 Filter 拦截，丢弃本次采集值
			if !thing.PostProcess(m, values) {
				continue
			}
		}

		// 轮询结果直接写入设备属性（历史数据由 BaseDevice 落库）
		_ = target.ReportProperties(values, time.Now())
	}
}
