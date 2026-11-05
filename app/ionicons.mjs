
import svgtofont from 'svgtofont';
import { JSDOM } from 'jsdom';
import { Worker } from 'worker_threads';
import fs from 'fs';
import path from 'path';

// ============ 字体配置 ============
// 同时处理 filled (svg/) 和 outline (svg-outline/) 两套图标
const FONTS = [
  //{ src: 'ionicons/svg', dist: 'ionicons', name: 'ionicons' },
  { src: 'ionicons/svg-sharp', dist: 'ionicons', name: 'ionicons' },
];

// ============ 样式解析 ============

/**
 * 解析 SVG 元素的 style 属性和直接属性，合并为统一对象
 */
function parseStyle(el) {
  const styleStr = el.getAttribute('style') || '';
  const style = {};
  styleStr.split(';').forEach(prop => {
    const idx = prop.indexOf(':');
    if (idx > 0) {
      const key = prop.slice(0, idx).trim();
      const val = prop.slice(idx + 1).trim();
      if (key) style[key] = val;
    }
  });
  // 直接属性覆盖 style 中的值
  for (const attr of ['fill', 'stroke', 'stroke-width', 'stroke-linecap', 'stroke-linejoin', 'stroke-miterlimit']) {
    const val = el.getAttribute(attr);
    if (val !== null) style[attr] = val;
  }
  return style;
}

/** 获取描边宽度（数值） */
function getStrokeWidth(style) {
  const sw = style['stroke-width'];
  if (!sw) return 0;
  return parseFloat(String(sw).replace(/px$/, '')) || 0;
}

/** 是否有描边 */
function hasStroke(style) {
  return !!style.stroke && style.stroke !== 'none';
}

/** 是否有填充（SVG 默认填充为黑色） */
function hasFill(style) {
  return !style.fill || style.fill !== 'none';
}

// ============ 形状转 path data ============

function circleToPath(cx, cy, r) {
  const d = 2 * r;
  return `M${cx - r},${cy} a${r},${r} 0 1,0 ${d},0 a${r},${r} 0 1,0 ${-d},0 Z`;
}

function ellipseToPath(cx, cy, rx, ry) {
  const dx = 2 * rx;
  return `M${cx - rx},${cy} a${rx},${ry} 0 1,0 ${dx},0 a${rx},${ry} 0 1,0 ${-dx},0 Z`;
}

function rectToPath(x, y, w, h, rx = 0, ry = 0) {
  if (rx <= 0 && ry <= 0) {
    return `M${x},${y} L${x + w},${y} L${x + w},${y + h} L${x},${y + h} Z`;
  }
  // 圆角矩形
  return `M${x + rx},${y} L${x + w - rx},${y} Q${x + w},${y} ${x + w},${y + ry} L${x + w},${y + h - ry} Q${x + w},${y + h} ${x + w - rx},${y + h} L${x + rx},${y + h} Q${x},${y + h} ${x},${y + h - ry} L${x},${y + ry} Q${x},${y} ${x + rx},${y} Z`;
}

function lineToPath(x1, y1, x2, y2) {
  return `M${x1},${y1} L${x2},${y2}`;
}

function pointsToPath(points, closed = false) {
  if (points.length < 4) return '';
  let d = `M${points[0]},${points[1]}`;
  for (let i = 2; i < points.length; i += 2) {
    d += ` L${points[i]},${points[i + 1]}`;
  }
  if (closed) d += ' Z';
  return d;
}

function pointsToArray(pointsStr) {
  return pointsStr.trim().split(/[\s,]+/).map(Number);
}

/**
 * 统计路径中的曲线命令数量（贝塞尔 + 圆弧），用于判断复杂度
 * 注意：SVG 路径中命令字母后跟多组坐标时会隐式重复该命令，
 * 例如 "A r1,r2,...,x1,y1 x2,y2" 实际是 2 个弧命令。
 * 这里按命令字母出现次数估算，已足够用于复杂度分级。
 */
function countCurves(pathData) {
  if (typeof pathData !== 'string') return 0;
  // 匹配所有曲线命令：A(圆弧) C/S(三次贝塞尔) Q/T(二次贝塞尔)
  const matches = pathData.match(/[ACQSTacqst]/g);
  return matches ? matches.length : 0;
}

// ============ Worker 超时保护 ============

/**
 * 在 worker 线程中运行 spo()，超时自动终止
 * 防止 construct.svg 等复杂路径导致 spo() 卡死整个进程
 */
function spoWithTimeout(spoInput, distance, options, timeoutMs = 10000) {
  return new Promise((resolve, reject) => {
    const workerCode = `
      const { parentPort, workerData } = require('worker_threads');
      try {
        const spo = require('svg-path-outline');
        const { spoInput, distance, options } = workerData;
        const result = spo(spoInput, distance, options);
        parentPort.postMessage({ success: true, result });
      } catch (e) {
        parentPort.postMessage({ success: false, error: e.message });
      }
    `;

    const worker = new Worker(workerCode, {
      eval: true,
      workerData: { spoInput, distance, options },
    });

    let done = false;
    const timeout = setTimeout(() => {
      if (!done) {
        done = true;
        worker.terminate();
        reject(new Error(`超时 (${timeoutMs}ms)`));
      }
    }, timeoutMs);

    worker.on('message', (msg) => {
      if (!done) {
        done = true;
        clearTimeout(timeout);
        if (msg.success) {
          resolve(msg.result);
        } else {
          reject(new Error(msg.error));
        }
      }
    });

    worker.on('error', (err) => {
      if (!done) {
        done = true;
        clearTimeout(timeout);
        reject(err);
      }
    });
  });
}

// ============ Worker 自检 ============

/**
 * 启动时自检 worker 能否正常加载 svg-path-outline。
 * 如果 worker 启动就卡住（makerjs 加载慢），提前报错而不是每个图标都卡。
 */
async function checkWorkerAvailable() {
  try {
    // 用一个最简单的路径测试，超时 15 秒（主要给 require 的时间）
    const result = await spoWithTimeout('M0,0 L100,0', 5, {
      joints: 0, inside: true, outside: true, tagName: 'path', bezierAccuracy: 5,
    }, 15000);
    if (!result) throw new Error('worker 返回空结果');
    console.log('✅ worker 自检通过（svg-path-outline 可用）');
  } catch (e) {
    console.error(`❌ worker 自检失败：${e.message}`);
    console.error('   请检查 svg-path-outline 和 makerjs 是否正常安装');
    throw e;
  }
}

// ============ 元素转换 ============

/**
 * 将单个 SVG 元素转换为填充 path 元素数组。
 *
 * - 有描边时：使用 svg-path-outline 将描边展开为填充轮廓
 *   ⚠️ 所有 spo() 调用都通过 worker 超时保护，防止 makerjs 在复杂路径上卡死
 *   ⚠️ 超时/失败时降级为用原始路径作为填充（图标仍能显示，只是没有描边轮廓）
 * - 有填充时：保留原始形状的填充 path
 *
 * 返回 null 表示未知元素类型（跳过），返回空数组表示不可见元素（删除）。
 */
async function convertElement(el, doc) {
  const style = parseStyle(el);
  const tagName = el.tagName.toLowerCase();
  const stroke = hasStroke(style);
  const fill = hasFill(style);
  const strokeWidth = getStrokeWidth(style);

  // 获取 path data 和 svg-path-outline 所需的输入类型
  let pathData = null;
  let spoInput = null;
  let spoTagName = 'path';

  switch (tagName) {
    case 'line': {
      const x1 = parseFloat(el.getAttribute('x1'));
      const y1 = parseFloat(el.getAttribute('y1'));
      const x2 = parseFloat(el.getAttribute('x2'));
      const y2 = parseFloat(el.getAttribute('y2'));
      pathData = lineToPath(x1, y1, x2, y2);
      spoInput = pathData;
      break;
    }
    case 'polyline': {
      const pts = pointsToArray(el.getAttribute('points'));
      pathData = pointsToPath(pts, false);
      spoInput = pts;
      spoTagName = 'polyline';
      break;
    }
    case 'polygon': {
      const pts = pointsToArray(el.getAttribute('points'));
      pathData = pointsToPath(pts, true);
      spoInput = pts;
      spoTagName = 'polygon';
      break;
    }
    case 'circle': {
      const cx = parseFloat(el.getAttribute('cx'));
      const cy = parseFloat(el.getAttribute('cy'));
      const r = parseFloat(el.getAttribute('r'));
      pathData = circleToPath(cx, cy, r);
      spoInput = pathData;
      break;
    }
    case 'ellipse': {
      const cx = parseFloat(el.getAttribute('cx'));
      const cy = parseFloat(el.getAttribute('cy'));
      const rx = parseFloat(el.getAttribute('rx'));
      const ry = parseFloat(el.getAttribute('ry'));
      pathData = ellipseToPath(cx, cy, rx, ry);
      spoInput = pathData;
      break;
    }
    case 'rect': {
      const x = parseFloat(el.getAttribute('x') || '0');
      const y = parseFloat(el.getAttribute('y') || '0');
      const w = parseFloat(el.getAttribute('width'));
      const h = parseFloat(el.getAttribute('height'));
      const rx = parseFloat(el.getAttribute('rx') || '0');
      const ry = parseFloat(el.getAttribute('ry') || '0');
      pathData = rectToPath(x, y, w, h, rx, ry);
      spoInput = pathData;
      break;
    }
    case 'path': {
      pathData = el.getAttribute('d');
      spoInput = pathData;
      break;
    }
    default:
      return null; // 未知元素，跳过
  }

  if (!pathData) return null;

  const newPaths = [];

  // 有描边：使用 svg-path-outline 将描边展开为填充路径
  // ⚠️ 所有 spo() 调用必须通过 spoWithTimeout() 走 worker：
  //    1. countCurves() 无法准确统计隐式重复命令（如 "A ... x1,y1 x2,y2" 是 2 个弧但只算 1）
  //    2. 即使曲线数量少，makerjs 在某些路径组合下仍可能卡死
  //    3. worker 可以超时强制 terminate，保证主流程不阻塞
  if (stroke && strokeWidth > 0) {
    try {
      const distance = strokeWidth / 2;
      const curveCount = typeof spoInput === 'string' ? countCurves(spoInput) : 0;
      // 根据曲线复杂度动态调整精度（曲线越多，精度越低，避免卡死）
      const bezierAccuracy = curveCount > 30 ? 25 : curveCount > 10 ? 15 : 8;
      // 超时时间也随复杂度增加：简单 8s，中等 12s，复杂 20s
      const timeoutMs = curveCount > 30 ? 20000 : curveCount > 10 ? 12000 : 8000;

      let outline = null;
      try {
        outline = await spoWithTimeout(spoInput, distance, {
          joints: 0,
          inside: true,
          outside: true,
          tagName: spoTagName,
          bezierAccuracy,
        }, timeoutMs);
      } catch (e) {
        console.error(`  ⚠️  spo 超时/失败 (${curveCount} curves, ${timeoutMs}ms): ${e.message}`);
        // 降级：用原始路径填充（非描边轮廓，但至少图标能显示）
      }

      if (outline) {
        const p = doc.createElement('path');
        p.setAttribute('d', outline);
        p.setAttribute('fill', '#000');
        newPaths.push(p);
      } else if (!fill) {
        // spo 无结果且无填充：用原始路径作为填充降级
        const p = doc.createElement('path');
        p.setAttribute('d', pathData);
        p.setAttribute('fill', '#000');
        newPaths.push(p);
      }
    } catch (e) {
      console.error(`  ⚠️  描边展开失败 (${tagName}): ${e.message}`);
      if (!fill) return newPaths;
    }
  }

  // 有填充：添加填充 path（与描边轮廓叠加，合并为完整图形）
  if (fill) {
    const p = doc.createElement('path');
    p.setAttribute('d', pathData);
    p.setAttribute('fill', '#000');
    newPaths.push(p);
  }

  return newPaths;
}

/**
 * 转换整个 SVG：
 * 1. 将所有形状元素（line/polyline/polygon/circle/ellipse/rect）转为 path
 * 2. 将所有带 stroke 的元素转为填充路径（描边展开）
 * 3. 清理 style 等可能干扰 svgtofont 的属性
 */
async function convertSvg(svgContent) {
  const dom = new JSDOM(svgContent);
  const doc = dom.window.document;
  const svg = doc.querySelector('svg');
  if (!svg) return svgContent;

  const elements = Array.from(
    svg.querySelectorAll('line, polyline, polygon, circle, ellipse, rect, path')
  );

  for (const el of elements) {
    // 纯填充 path（无描边）：只清理属性，保持不变
    if (el.tagName.toLowerCase() === 'path') {
      const style = parseStyle(el);
      if (hasFill(style) && !hasStroke(style)) {
        el.setAttribute('fill', '#000');
        el.removeAttribute('style');
        el.removeAttribute('stroke');
        el.removeAttribute('stroke-width');
        continue;
      }
    }

    const newPaths = await convertElement(el, doc);
    if (newPaths === null) continue; // 未知元素，跳过

    // 用新 path 替换原元素（newPaths 为空时等于删除不可见元素）
    for (const p of newPaths) {
      el.parentNode.insertBefore(p, el);
    }
    el.parentNode.removeChild(el);
  }

  return svg.outerHTML;
}

// ============ 单套字体生成 ============

/**
 * 处理一套字体：转换 SVG → 生成字体文件
 * @param {{src: string, dist: string, name: string}} font
 */
async function generateFont(font) {
  const srcDir = path.resolve(font.src);
  const tempDir = path.resolve(font.src + '-converted');
  const distDir = path.resolve(font.dist);

  if (!fs.existsSync(srcDir)) {
    console.log(`\n⏭️  跳过 ${font.name}：源目录不存在 ${font.src}`);
    return;
  }

  console.log(`\n================ ${font.name} ================`);
  console.log(`🔄 转换 SVG（描边→填充，形状→path）...`);

  // 创建临时目录
  if (fs.existsSync(tempDir)) {
    fs.rmSync(tempDir, { recursive: true });
  }
  fs.mkdirSync(tempDir, { recursive: true });

  const files = fs.readdirSync(srcDir).filter(f => f.endsWith('.svg'));
  let ok = 0;
  let fail = 0;

  for (const file of files) {
    const start = Date.now();
    const src = path.join(srcDir, file);
    const dest = path.join(tempDir, file);
    const content = fs.readFileSync(src, 'utf-8');

    // 统计 path data 中的曲线数量（不含 style 属性中的字母）
    const pathDataList = content.match(/d="([^"]+)"/g) || [];
    const totalCurves = pathDataList.reduce((sum, m) => sum + countCurves(m), 0);
    if (totalCurves > 20) {
      console.log(`  📄 ${file} (${totalCurves} curves)...`);
    }

    try {
      fs.writeFileSync(dest, await convertSvg(content));
      ok++;
      const elapsed = Date.now() - start;
      if (elapsed > 200) {
        console.log(`  ⏱️  ${file}: ${elapsed}ms`);
      }
    } catch (e) {
      console.error(`  ⚠️  ${file}: ${e.message}`);
      fs.copyFileSync(src, dest); // 降级：使用原始文件
      fail++;
    }
  }
  console.log(`✅ 转换完成：${ok} 成功，${fail} 降级`);

  console.log(`🔄 生成字体 ${font.name}...`);

  await svgtofont({
    src: tempDir,
    dist: distDir,
    fontName: font.name,
    css: true,
    generateInfoData: true,
    startUnicode: 0xea01,

    svgicons2svgfont: {
      // 标准字体高度
      fontHeight: 1000,
      // 关闭 normalize —— 对复杂曲线计算包围盒会卡死
      // 所有图标 viewBox 都是 0 0 512 512，尺寸一致，不需要 normalize
      normalize: false,
      // 关闭 centerHorizontally —— 同样依赖包围盒计算，有卡死风险
      // ionicons 图标本身在 viewBox 中已水平居中
      centerHorizontally: false,
      // 不使用垂直居中 —— 会导致曲线超出 em 方块边界，产生阴影/黑块
      centerVertically: false,
    },

    website: {
      title: font.name,
      meta: {
        description: `${font.name} as icon font`,
        keywords: `iconfont,${font.name}`,
      },
    },
  });

  console.log(`🎉 ${font.name} 生成成功！`);

  // 清理临时目录
  fs.rmSync(tempDir, { recursive: true });
  console.log(`🧹 已清理临时目录`);
}

// ============ 主函数 ============

async function main() {
  // 启动前自检 worker，避免每个图标都卡在 worker 加载
  await checkWorkerAvailable();

  for (const font of FONTS) {
    await generateFont(font);
  }
  console.log('\n✨ 全部字体生成完成');
}

main().catch(err => {
  console.error('❌ 生成失败：', err);
  process.exit(1);
});
