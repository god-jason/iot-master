/**
 * PointsEditTool —— 顶点编辑工具（折线 / 曲线 / 管道流动线）
 *
 * 支持两种编辑模式：
 *  1. points  模式：Line + points（折线 / 管道流动线），顶点 = 折角点
 *  2. bezier  模式：Path + C 命令（曲线），锚点 + 贝塞尔控制手柄，类似 Photoshop
 *
 * 拖拽机制完全仿照 Leafer 官方 LineEditTool，使用 EditBox 自带的 resize 拖拽事件流：
 *   拖拽自定义 EditPoint
 *     → EditBox.onDragStart(e)  设置 resizing = true
 *     → EditBox.onDrag(e)       调用 transformTool.onScale(e)
 *     → TransformTool.onScale   检测到 editTool.onScaleWithDrag 存在
 *     → TransformTool.scaleWithDrag(data)
 *     → PointsEditTool.onScaleWithDrag(e)   ← 我们在这里更新顶点
 *
 * 关键坐标 API（Leafer UIElement）：
 *   innerToWorld(inner, to, distance, relative)  → 原地修改 to（返回 undefined）
 *     把 inner 由 worldTransform 变换到世界坐标，再反变换进 relative(editBox) 坐标系
 *   drag.getInnerMove(target)  → 拖动增量（每帧），在 target 本地坐标系下的分量（含缩放换算）
 *
 * 顶点增删：
 *   - 段中点橙色 "+" 手柄：tap 插入新顶点（贝塞尔在 t=0.5 用 de Casteljau 二分）
 *   - 顶点/锚点手柄：double_tap 删除该顶点（少于 2 个顶点时禁止删除）
 */

import { EditToolCreator, EditTool, EditPoint } from '@leafer-in/editor'
import { Line } from 'leafer-editor'

// 事件名直接写字符串，与 DragEvent.START/DRAG/END 一致，避免不同包 PointerEvent 命名冲突
const DRAG_START = 'drag.start'
const DRAG = 'drag'
const DRAG_END = 'drag.end'
const TAP = 'tap'
const DOUBLE_TAP = 'double_tap'

// Leafer SVG path 命令码，与 @leafer/path 的 PathCommandMap 一致
// M = 1 (moveto, 3 个数字) / L = 2 (lineto, 3 个数字) / C = 5 (curveto, 7 个数字) / Z = 11
const M = 1, C = 5

// 各类手柄的样式
const ANCHOR_STYLE = { around: 'center', width: 12, height: 12, cornerRadius: 2, fill: '#ffffff', stroke: '#1890ff', strokeWidth: 2, hitFill: 'all', hitRadius: 12, cursor: 'move' }
const CONTROL_STYLE = { around: 'center', width: 10, height: 10, cornerRadius: 5, fill: '#ffffff', stroke: '#00c853', strokeWidth: 2, hitFill: 'all', hitRadius: 10, cursor: 'move' }
const ADD_STYLE = { around: 'center', width: 10, height: 10, cornerRadius: 5, fill: '#ffffff', stroke: '#ff9800', strokeWidth: 2, hitFill: 'all', hitRadius: 10, cursor: 'pointer' }

// 点线性插值 / 三次贝塞尔取点
const lerp = (p, q, t) => ({ x: p.x + (q.x - p.x) * t, y: p.y + (q.y - p.y) * t })

export class PointsEditTool extends EditTool {

  get tag() { return 'PointsEditTool' }

  constructor(editor) {
    super(editor)
    this._anchors = []    // 统一顶点模型：[{ x, y, hLeft?, hRight? }]
    this._mode = 'points' // 'points' | 'bezier'
    this._handles = []    // 可拖拽手柄（顶点 / 锚点 / 控制点）
    this._adds = []       // 段中点 "+" 插入手柄
    this._builtMode = null
    this._builtCount = -1
  }

  // ==================== 生命周期 ====================

  onLoad() {
    this._syncHandles()
  }

  onUpdate() {
    this._hideChrome()
    this._syncHandles()
  }

  onUnload() {
    this._clearAll()
    super.onUnload && super.onUnload()
  }

  // ---- 隐藏 EditBox 默认控件（顶点编辑模式下不需要矩形外框/缩放/旋转控件） ----

  _hideChrome() {
    const eb = this.editBox
    if (!eb) return
    const list = [
      eb.rect, eb.circle, eb.buttons,
      ...(eb.resizePoints || []), ...(eb.rotatePoints || []), ...(eb.resizeLines || [])
    ]
    list.forEach(el => { if (el) { el.visible = 0; el.hittable = false } })
  }

  // ==================== 顶点模型：解析 / 回写 ====================

  _loadModel() {
    const target = this.editBox.target
    if (!target) return
    if (Array.isArray(target.points)) {
      this._mode = 'points'
      const pts = target.points
      const arr = []
      for (let i = 0; i + 1 < pts.length; i += 2) arr.push({ x: pts[i], y: pts[i + 1] })
      this._anchors = arr
    } else {
      this._mode = 'bezier'
      this._anchors = this._parseBezier(target)
    }
  }

  // 解析 Path 命令数组为锚点模型（含左右控制手柄）
  _parseBezier(target) {
    const data = target.getPath ? target.getPath() : (target.path || [])
    const arr = []
    let i = 0, len = data.length, prev
    while (i < len) {
      const cmd = data[i]
      if (cmd === M || cmd === 2 /* L */) {
        arr.push(prev = { x: data[i + 1], y: data[i + 2], hLeft: null, hRight: null })
        i += 3
      } else if (cmd === C) {
        if (prev) prev.hRight = { x: data[i + 1], y: data[i + 2] }
        arr.push(prev = { x: data[i + 5], y: data[i + 6], hLeft: { x: data[i + 3], y: data[i + 4] }, hRight: null })
        i += 7
      } else if (cmd === 11 /* Z */) {
        i += 1
      } else {
        i += 1 // 未知命令保守跳过
      }
    }
    return arr
  }

  // 把顶点模型序列化回目标图元
  _writeBack() {
    const target = this.editBox.target
    const anchors = this._anchors
    if (!target || !anchors.length) return

    if (this._mode === 'points') {
      const flat = []
      anchors.forEach(p => { flat.push(p.x, p.y) })
      target.points = flat
    } else {
      const data = [M, anchors[0].x, anchors[0].y]
      for (let i = 0; i < anchors.length - 1; i++) {
        const a = anchors[i], b = anchors[i + 1]
        const c1 = a.hRight || a
        const c2 = b.hLeft || b
        data.push(C, c1.x, c1.y, c2.x, c2.y, b.x, b.y)
      }
      target.path = data
    }
  }

  // ==================== 拖拽：更新顶点 ====================

  /**
   * 拖拽某个手柄时的回调（由 TransformTool.scaleWithDrag 调用）。
   * e.drag 是 DragEvent；e.drag.current 是正在拖拽的 EditPoint。
   */
  onScaleWithDrag(e) {
    const point = e.drag && e.drag.current
    const meta = point && point.__ptsMeta
    if (!meta) return

    const move = e.drag.getInnerMove(e.target) // 增量（每帧）
    this._applyMove(meta, move.x, move.y)
    this._writeBack()
    this._positionAll()
  }

  _applyMove(meta, dx, dy) {
    const anchors = this._anchors
    if (meta.kind === 'anchor') {
      // 移动锚点：控制手柄跟随一起移动（保持相对位置）
      const a = anchors[meta.anchorIndex]
      if (!a) return
      a.x += dx; a.y += dy
      if (a.hLeft) { a.hLeft.x += dx; a.hLeft.y += dy }
      if (a.hRight) { a.hRight.x += dx; a.hRight.y += dy }
    } else if (meta.kind === 'control') {
      // 移动控制手柄：另一侧手柄反向镜像，保持平滑（类似 Photoshop 钢笔）
      const a = anchors[meta.anchorIndex]
      if (!a) return
      const c = meta.side === 'left' ? a.hLeft : a.hRight
      if (!c) return
      c.x += dx; c.y += dy
      const mirror = meta.side === 'left' ? a.hRight : a.hLeft
      if (mirror) { mirror.x -= dx; mirror.y -= dy }
    }
  }

  // ==================== 顶点增删 ====================

  _insertVertex(segIndex) {
    const anchors = this._anchors
    if (segIndex < 0 || segIndex + 1 >= anchors.length) return

    if (this._mode === 'points') {
      const a = anchors[segIndex], b = anchors[segIndex + 1]
      anchors.splice(segIndex + 1, 0, { x: (a.x + b.x) / 2, y: (a.y + b.y) / 2 })
    } else {
      // 三次贝塞尔 t=0.5 处用 de Casteljau 二分，插入平滑新锚点
      const a = anchors[segIndex], b = anchors[segIndex + 1]
      const p0 = a, p3 = b
      const p1 = a.hRight || a, p2 = b.hLeft || b
      const ab = lerp(p0, p1, 0.5), bc = lerp(p1, p2, 0.5), cd = lerp(p2, p3, 0.5)
      const abc = lerp(ab, bc, 0.5), bcd = lerp(bc, cd, 0.5)
      const abcd = lerp(abc, bcd, 0.5) // 曲线上新锚点

      a.hRight = ab
      b.hLeft = cd
      anchors.splice(segIndex + 1, 0, { x: abcd.x, y: abcd.y, hLeft: abc, hRight: bcd })
    }

    this._writeBack()
    this._syncHandles()
  }

  _deleteVertex(index) {
    const anchors = this._anchors
    if (anchors.length <= 2) return // 至少保留两个顶点

    if (this._mode === 'points') {
      anchors.splice(index, 1)
    } else {
      if (index === 0) {
        anchors.splice(0, 1)
        if (anchors[0]) anchors[0].hLeft = null
      } else if (index === anchors.length - 1) {
        anchors.pop()
        if (anchors[anchors.length - 1]) anchors[anchors.length - 1].hRight = null
      } else {
        anchors.splice(index, 1)
      }
    }

    this._writeBack()
    this._syncHandles()
  }

  // ==================== 手柄渲染 ====================

  _syncHandles() {
    const target = this.editBox.target
    if (!target) return
    this._loadModel()
    const n = this._anchors.length

    if (this._builtMode !== this._mode || this._builtCount !== n) {
      this._clearAll()
      this._buildVisuals()
      this._builtMode = this._mode
      this._builtCount = n
    } else {
      this._positionAll()
    }
  }

  _buildVisuals() {
    const anchors = this._anchors
    const n = anchors.length

    // 1. 顶点 / 锚点手柄（可拖拽 + 可删除）
    for (let i = 0; i < n; i++) {
      this._createHandle({ kind: 'anchor', anchorIndex: i }, ANCHOR_STYLE, true)
    }

    // 2. 段中点 "+" 插入手柄
    for (let i = 0; i < n - 1; i++) {
      this._createAdd(i)
    }

    // 3. 贝塞尔控制手柄 + 连线
    if (this._mode === 'bezier') {
      for (let i = 0; i < n; i++) {
        const a = anchors[i]
        if (a.hLeft) this._createHandle({ kind: 'control', anchorIndex: i, side: 'left' }, CONTROL_STYLE, false)
        if (a.hRight) this._createHandle({ kind: 'control', anchorIndex: i, side: 'right' }, CONTROL_STYLE, false)
      }
    }

    this._positionAll()
  }

  _createHandle(meta, style, deletable) {
    const p = new EditPoint({ ...style })
    p.pointType = 'resize'  // 触发 EditBox 的 resize 拖拽流
    p.direction = 7         // 一个合法的 resize 方向，仅用于 getScaleData 分支（结果被忽略）
    p.__ptsMeta = meta      // 私有标记，用于 onScaleWithDrag 定位

    const eb = this.editBox
    const eventIds = [p.on_([
      [DRAG_START, eb.onDragStart, eb],
      [DRAG, eb.onDrag, eb],
      [DRAG_END, eb.onDragEnd, eb],
    ])]

    if (deletable) {
      eventIds.push(p.on_(DOUBLE_TAP, () => this._deleteVertex(meta.anchorIndex)))
    }

    let lineEl = null
    if (meta.kind === 'control') {
      // 锚点 → 控制手柄的连线
      lineEl = new Line({ name: 'pts-control-line', stroke: '#909399', strokeWidth: 1, dashPattern: [4, 3], hittable: false })
      eb.view.add(lineEl)
    }

    eb.view.add(p)
    this._handles.push({ pointEl: p, meta, eventIds, lineEl })
  }

  _createAdd(segIndex) {
    const p = new EditPoint({ ...ADD_STYLE })
    p.__segIndex = segIndex
    const tapId = p.on_(TAP, () => this._insertVertex(segIndex))
    this.editBox.view.add(p)
    this._adds.push({ pointEl: p, segIndex, tapId })
  }

  _positionAll() {
    const anchors = this._anchors
    const eb = this.editBox
    const target = eb.target
    if (!target) return

    const toEB = (x, y) => {
      const p = { x, y }
      target.innerToWorld(p, p, false, eb)
      return p
    }

    this._handles.forEach(h => {
      const m = h.meta
      let px, py

      if (m.kind === 'anchor') {
        const a = anchors[m.anchorIndex]
        if (!a) return
        px = a.x; py = a.y
      } else {
        const a = anchors[m.anchorIndex]
        const c = m.side === 'left' ? a.hLeft : a.hRight
        if (!c) return
        px = c.x; py = c.y
      }

      const e = toEB(px, py)
      h.pointEl.set({ x: e.x, y: e.y })

      if (h.lineEl && m.kind === 'control') {
        const a = anchors[m.anchorIndex]
        const ae = toEB(a.x, a.y)
        h.lineEl.set({ x: ae.x, y: ae.y, toPoint: { x: e.x - ae.x, y: e.y - ae.y } })
      }
    })

    this._adds.forEach(ad => {
      const mid = this._segmentMidpoint(ad.segIndex)
      if (!mid) return
      const e = toEB(mid.x, mid.y)
      ad.pointEl.set({ x: e.x, y: e.y })
    })
  }

  // 段中点：折线取线段中点；贝塞尔取曲线上 t=0.5 的点
  _segmentMidpoint(i) {
    const a = this._anchors[i], b = this._anchors[i + 1]
    if (!a || !b) return null
    if (this._mode !== 'bezier') {
      return { x: (a.x + b.x) / 2, y: (a.y + b.y) / 2 }
    }
    const p0 = a, p3 = b
    const p1 = a.hRight || a, p2 = b.hLeft || b
    return this._cubicPoint(p0, p1, p2, p3, 0.5)
  }

  _cubicPoint(p0, p1, p2, p3, t) {
    const u = 1 - t
    const uu = u * u, tt = t * t, uuu = uu * u, ttt = tt * t
    return {
      x: uuu * p0.x + 3 * uu * t * p1.x + 3 * u * tt * p2.x + ttt * p3.x,
      y: uuu * p0.y + 3 * uu * t * p1.y + 3 * u * tt * p2.y + ttt * p3.y
    }
  }

  // ==================== 清理 ====================

  _clearAll() {
    this._handles.forEach(h => {
      h.eventIds.forEach(id => { try { h.pointEl.off_(id) } catch (_) {} })
      if (h.lineEl) { try { h.lineEl.remove() } catch (_) {} }
      try { h.pointEl.remove() } catch (_) {}
    })
    this._adds.forEach(ad => {
      try { ad.pointEl.off_(ad.tapId) } catch (_) {}
      try { ad.pointEl.remove() } catch (_) {}
    })
    this._handles = []
    this._adds = []
  }
}

EditToolCreator.register(PointsEditTool)