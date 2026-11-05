// 文件上传管理接口
// 提供图片、附件、头像、图标等文件的上传处理能力
// 支持图片缩放、格式转换、质量压缩，上传目录从全局配置读取
package api

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/god-jason/iot-master/pkg/config"
	"github.com/god-jason/iot-master/pkg/curd"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
)

// UploadHandler 文件上传处理器
// 统一管理上传目录的初始化和各类文件上传逻辑
// 支持的上传类型：普通图片、附件、用户头像、应用图标、原始图片
type UploadHandler struct {
	uploadDir string // 上传根目录路径，从配置读取或默认为./uploads
}

// NewUploadHandler 创建文件上传处理器实例
// 功能: 初始化上传根目录，自动创建images/attachments/temp三个子目录（权限0755）
// 返回: 初始化后的UploadHandler指针
func NewUploadHandler() *UploadHandler {
	uploadDir := config.GetString("upload", "dir")
	if uploadDir == "" {
		uploadDir = "./uploads" // 未配置时使用默认目录
	}
	// 自动创建所需的子目录，不存在则递归创建
	os.MkdirAll(filepath.Join(uploadDir, "images"), 0755)      // 图片存储目录
	os.MkdirAll(filepath.Join(uploadDir, "attachments"), 0755) // 附件存储目录
	os.MkdirAll(filepath.Join(uploadDir, "temp"), 0755)        // 临时文件目录
	return &UploadHandler{uploadDir: uploadDir}
}

// ImageUploadRequest 图片上传参数结构体
// 通过URL Query参数传递，控制图片处理行为
type ImageUploadRequest struct {
	Width   int    `form:"width"`   // 目标图片宽度（像素），0表示保持原图宽度
	Height  int    `form:"height"`  // 目标图片高度（像素），0表示保持原图高度
	Format  string `form:"format"`  // 输出图片格式：jpg/jpeg/png/webp，默认jpg
	Quality int    `form:"quality"` // JPEG压缩质量 1-100，数值越高质量越好文件越大，默认85
}

// getServerURL 获取上传文件的访问URL前缀
// 参数: c - Gin请求上下文，用于判断HTTP/HTTPS和读取Host
// 返回: 服务器基础URL，如 https://example.com 或配置的固定ServerURL
func getServerURL(c *gin.Context) string {
	// 优先使用配置文件中指定的固定访问域名
	serverURL := config.GetString("upload", "server_url")
	if serverURL != "" {
		return serverURL
	}
	// 未配置时根据请求自动推断协议和域名
	proto := "http"
	if c.Request.TLS != nil {
		proto = "https" // TLS连接则使用https
	}
	return fmt.Sprintf("%s://%s", proto, c.Request.Host)
}

// buildFullURL 拼接完整的文件访问URL
// 参数:
//
//	serverURL - 服务器基础URL（可空，表示相对路径）
//	relativePath - 文件的相对路径，如 /images/xxx.jpg
//
// 返回: 完整URL或规范化的相对路径
func buildFullURL(serverURL, relativePath string) string {
	if serverURL == "" {
		return relativePath // 无配置时返回相对路径
	}
	// 去除serverURL末尾斜杠，与relativePath拼接避免重复斜杠
	serverURL = strings.TrimSuffix(serverURL, "/")
	return serverURL + relativePath
}

// UploadImage 上传并处理图片（支持缩放、格式转换、质量调节）
// 路由: POST /upload/image，form-data字段名: file
// Query参数: width(宽)/height(高)/format(格式)/quality(质量)
// 返回: { url: 访问URL, filename: 文件名, format: 格式, width: 实际宽, height: 实际高 }
// 限制: 仅支持图片格式，单文件最大10MB
// 错误处理: 400非图片/文件过大/无法解析图片, 500读写文件失败
func (h *UploadHandler) UploadImage(c *gin.Context) {
	var req ImageUploadRequest
	// 从Query参数读取并解析图片处理参数
	req.Width, _ = strconv.Atoi(c.DefaultQuery("width", "0"))
	req.Height, _ = strconv.Atoi(c.DefaultQuery("height", "0"))
	req.Format = c.DefaultQuery("format", "jpg")
	quality, _ := strconv.Atoi(c.DefaultQuery("quality", "85"))
	// 质量参数范围校验，超出范围回退默认值85
	if quality < 1 || quality > 100 {
		quality = 85
	}
	req.Quality = quality

	// 尺寸参数校验：拒绝负数，并限制单边与像素总量，避免图片炸弹耗尽内存
	if req.Width < 0 || req.Height < 0 {
		curd.Error(c, 400, "图片尺寸不能为负数")
		return
	}
	if req.Width > 4096 || req.Height > 4096 {
		curd.Error(c, 400, "图片宽高不能超过4096像素")
		return
	}
	if req.Width > 0 && req.Height > 0 && req.Width*req.Height > 4096*4096 {
		curd.Error(c, 400, "图片像素总量过大")
		return
	}

	// 输出格式白名单，仅支持 png/jpg/jpeg，其余统一按 jpg 处理
	format := strings.ToLower(req.Format)
	switch format {
	case "png", "jpg", "jpeg":
	default:
		format = "jpg"
	}
	req.Format = format

	// 读取上传的form-data文件
	file, err := c.FormFile("file")
	if err != nil {
		curd.Error(c, 400, "请选择要上传的文件")
		return
	}

	// 校验文件扩展名是否为支持的图片类型
	if !isImageFile(file.Filename) {
		curd.Error(c, 400, "只支持上传图片文件")
		return
	}

	// 校验文件大小不超过10MB
	if file.Size > 10*1024*1024 {
		curd.Error(c, 400, "文件大小不能超过10MB")
		return
	}

	// 打开上传文件的读取流
	src, err := file.Open()
	if err != nil {
		curd.Error(c, 500, "无法读取文件")
		return
	}
	defer src.Close()

	// 解码图片为image.Image对象（自动识别jpeg/png/gif等格式）
	img, _, err := image.Decode(src)
	if err != nil {
		curd.Error(c, 400, "无法解析图片格式")
		return
	}

	// 根据目标格式确定文件扩展名（扩展名与实际编码格式保持一致）
	ext := getExtFromFormat(req.Format)
	// 生成唯一文件名：8位UUID前缀 + 时间戳 + 扩展名，避免冲突
	filename := fmt.Sprintf("%s_%d%s", uuid.New().String()[:8], time.Now().Unix(), ext)
	relativePath := fmt.Sprintf("/images/%s", filename)        // 相对访问路径
	fullPath := filepath.Join(h.uploadDir, "images", filename) // 磁盘绝对路径

	// 按指定尺寸缩放图片，Bicubic算法保证缩放质量
	var processedImg image.Image
	if req.Width > 0 || req.Height > 0 {
		// nfnt/resize 在单边为 0 时按原图比例自适应另一边
		processedImg = resize.Resize(uint(req.Width), uint(req.Height), img, resize.Bicubic)
	} else {
		processedImg = img // 无尺寸要求则使用原图
	}

	// 创建目标文件用于写入
	out, err := os.Create(fullPath)
	if err != nil {
		curd.Error(c, 500, "无法创建文件")
		return
	}
	defer out.Close()

	// 根据格式选择编码器写入
	switch req.Format {
	case "png":
		err = png.Encode(out, processedImg) // PNG无损编码
	default:
		// jpg/jpeg 使用 JPEG 有损压缩
		err = jpeg.Encode(out, processedImg, &jpeg.Options{Quality: req.Quality})
	}

	if err != nil {
		curd.Error(c, 500, "无法保存图片")
		return
	}

	// 返回上传结果
	curd.Success(c, gin.H{
		"url":      buildFullURL(getServerURL(c), relativePath), // 完整访问URL
		"filename": filename,                                    // 生成的文件名
		"format":   req.Format,                                  // 实际输出格式
		"width":    processedImg.Bounds().Dx(),                  // 处理后实际宽度
		"height":   processedImg.Bounds().Dy(),                  // 处理后实际高度
	})
}

// UploadAttachment 上传附件（不做处理，原样保存）
// 路由: POST /upload/attachment，form-data字段名: file
// 返回: { url: 访问URL, filename: 文件名, size: 文件大小字节, ext: 扩展名 }
// 限制: 单文件最大50MB
func (h *UploadHandler) UploadAttachment(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		curd.Error(c, 400, "请选择要上传的文件")
		return
	}

	// 附件大小限制50MB
	if file.Size > 50*1024*1024 {
		curd.Error(c, 400, "文件大小不能超过50MB")
		return
	}

	// 扩展名白名单：禁止 html/svg/js 等可被浏览器执行的文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedAttachmentExt(ext) {
		curd.Error(c, 400, "不支持的文件类型")
		return
	}

	// 保留原文件扩展名，文件名仍用UUID+时间戳确保唯一
	filename := fmt.Sprintf("%s_%d%s", uuid.New().String()[:8], time.Now().Unix(), ext)
	relativePath := fmt.Sprintf("/attachments/%s", filename)
	fullPath := filepath.Join(h.uploadDir, "attachments", filename)

	// Gin提供的便捷方法直接保存上传文件
	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		curd.Error(c, 500, "无法保存文件")
		return
	}

	curd.Success(c, gin.H{
		"url":      buildFullURL(getServerURL(c), relativePath),
		"filename": filename,
		"size":     file.Size, // 文件原始大小（字节）
		"ext":      ext,       // 文件扩展名（含点号，如.pdf）
	})
}

// UploadAvatar 上传用户头像（自动裁剪为200x200正方形JPG）
// 路由: POST /upload/avatar，form-data字段名: file
// 返回: { url: 头像访问URL }
// 限制: 仅支持图片格式，单文件最大5MB
func (h *UploadHandler) UploadAvatar(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		curd.Error(c, 400, "请选择要上传的文件")
		return
	}

	if !isImageFile(file.Filename) {
		curd.Error(c, 400, "只支持上传图片文件")
		return
	}

	// 头像文件限制5MB
	if file.Size > 5*1024*1024 {
		curd.Error(c, 400, "文件大小不能超过5MB")
		return
	}

	src, err := file.Open()
	if err != nil {
		curd.Error(c, 500, "无法读取文件")
		return
	}
	defer src.Close()

	img, _, err := image.Decode(src)
	if err != nil {
		curd.Error(c, 400, "无法解析图片格式")
		return
	}

	// 裁剪为正方形并缩放到200x200像素（头像标准尺寸）
	squared := squareImage(img, 200)
	if squared == nil {
		squared = img // 裁剪失败兜底使用原图
	}

	// 头像统一使用jpg格式，文件名带avatar_前缀
	filename := fmt.Sprintf("avatar_%s_%d.jpg", uuid.New().String()[:8], time.Now().Unix())
	relativePath := fmt.Sprintf("/images/%s", filename)
	fullPath := filepath.Join(h.uploadDir, "images", filename)

	out, err := os.Create(fullPath)
	if err != nil {
		curd.Error(c, 500, "无法创建文件")
		return
	}
	defer out.Close()

	// 固定质量85的JPEG编码
	err = jpeg.Encode(out, squared, &jpeg.Options{Quality: 85})
	if err != nil {
		curd.Error(c, 500, "无法保存图片")
		return
	}

	curd.Success(c, gin.H{
		"url": buildFullURL(getServerURL(c), relativePath),
	})
}

// UploadIcon 上传应用/功能图标（自动裁剪为128x128正方形PNG）
// 路由: POST /upload/icon，form-data字段名: file
// 返回: { url: 图标访问URL }
// 限制: 仅支持图片格式，单文件最大2MB
func (h *UploadHandler) UploadIcon(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		curd.Error(c, 400, "请选择要上传的文件")
		return
	}

	if !isImageFile(file.Filename) {
		curd.Error(c, 400, "只支持上传图片文件")
		return
	}

	// 图标文件限制2MB
	if file.Size > 2*1024*1024 {
		curd.Error(c, 400, "文件大小不能超过2MB")
		return
	}

	src, err := file.Open()
	if err != nil {
		curd.Error(c, 500, "无法读取文件")
		return
	}
	defer src.Close()

	img, _, err := image.Decode(src)
	if err != nil {
		curd.Error(c, 400, "无法解析图片格式")
		return
	}

	// 裁剪为128x128正方形；裁剪失败则直接缩放到128x128
	squared := squareImage(img, 128)
	if squared == nil {
		squared = resize.Resize(128, 128, img, resize.Bicubic)
	}

	// 图标统一使用PNG格式（支持透明通道），文件名带icon_前缀
	filename := fmt.Sprintf("icon_%s_%d.png", uuid.New().String()[:8], time.Now().Unix())
	relativePath := fmt.Sprintf("/images/%s", filename)
	fullPath := filepath.Join(h.uploadDir, "images", filename)

	out, err := os.Create(fullPath)
	if err != nil {
		curd.Error(c, 500, "无法创建文件")
		return
	}
	defer out.Close()

	err = png.Encode(out, squared) // PNG无损编码
	if err != nil {
		curd.Error(c, 500, "无法保存图片")
		return
	}

	curd.Success(c, gin.H{
		"url": buildFullURL(getServerURL(c), relativePath),
	})
}

// UploadRaw 上传原始图片（不做任何缩放或格式转换，原样保存）
// 路由: POST /upload/raw，form-data字段名: file
// 返回: { url: 访问URL, filename: 文件名, size: 文件大小字节 }
// 限制: 仅支持图片格式，单文件最大10MB
func (h *UploadHandler) UploadRaw(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		curd.Error(c, 400, "请选择要上传的文件")
		return
	}

	if !isImageFile(file.Filename) {
		curd.Error(c, 400, "只支持上传图片文件")
		return
	}

	if file.Size > 10*1024*1024 {
		curd.Error(c, 400, "文件大小不能超过10MB")
		return
	}

	// 保留原始扩展名
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s_%d%s", uuid.New().String()[:8], time.Now().Unix(), ext)
	relativePath := fmt.Sprintf("/images/%s", filename)
	fullPath := filepath.Join(h.uploadDir, "images", filename)

	// 直接保存不处理
	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		curd.Error(c, 500, "无法保存文件")
		return
	}

	curd.Success(c, gin.H{
		"url":      buildFullURL(getServerURL(c), relativePath),
		"filename": filename,
		"size":     file.Size,
	})
}

// DeleteFile 删除已上传的文件
// 路由: DELETE /upload/file?path=xxx
// Query参数: path - 文件相对路径，如 /images/xxx.jpg
// 返回: 无
// 错误处理: 400缺少path, 404文件不存在, 500删除失败
func (h *UploadHandler) DeleteFile(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		curd.Error(c, 400, "请提供文件路径")
		return
	}

	// 兼容上传接口返回的完整访问URL（含协议与域名），提取其中的路径部分
	if u, err := url.Parse(path); err == nil && u.Scheme != "" && u.Host != "" {
		path = u.Path
	}

	// 规范化路径并校验必须位于上传根目录内，防止 ../ 路径穿越删除任意文件
	rel := strings.TrimPrefix(filepath.ToSlash(filepath.Clean("/"+path)), "/")
	fullPath := filepath.Join(h.uploadDir, filepath.FromSlash(rel))
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		curd.Error(c, 400, "无效的文件路径")
		return
	}
	absDir, err := filepath.Abs(h.uploadDir)
	if err != nil {
		curd.Error(c, 500, "上传目录不可用")
		return
	}
	if !strings.HasPrefix(absPath, absDir+string(os.PathSeparator)) {
		curd.Error(c, 400, "无效的文件路径")
		return
	}
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		curd.Error(c, 404, "文件不存在")
		return
	}

	// 执行删除
	if err := os.Remove(absPath); err != nil {
		curd.Error(c, 500, "删除失败")
		return
	}

	curd.Success(c, nil)
}

// ========== 以下为内部辅助函数 ==========

// isImageFile 判断文件名是否为支持的图片格式（通过扩展名）
// 参数: filename - 文件名或完整路径
// 返回: true表示是支持的图片格式（jpg/jpeg/png/gif/bmp/webp）
func isImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
	for _, e := range imageExts {
		if ext == e {
			return true
		}
	}
	return false
}

// allowedAttachmentExt 判断附件扩展名是否在白名单内（仅允许常见文档/压缩包）
// 参数: ext - 小写扩展名（含点号）
// 返回: true 表示允许上传
func allowedAttachmentExt(ext string) bool {
	allowed := map[string]bool{
		".pdf": true, ".doc": true, ".docx": true,
		".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true,
		".txt": true, ".csv": true, ".md": true, ".json": true, ".xml": true,
		".zip": true, ".rar": true, ".7z": true, ".tar": true, ".gz": true,
	}
	return allowed[ext]
}

// getExtFromFormat 根据格式名称返回对应的文件扩展名（带点号）
// 参数: format - 格式标识（png/jpg/jpeg）
// 返回: 带点号的扩展名，如 .jpg .png
func getExtFromFormat(format string) string {
	switch strings.ToLower(format) {
	case "png":
		return ".png"
	default:
		return ".jpg" // jpg/jpeg 及未知格式统一使用 JPEG 扩展名
	}
}

// squareImage 将图片裁剪居中为正方形并缩放到指定尺寸
// 参数:
//
//	img - 源图片
//	size - 目标正方形边长（像素）
//
// 返回: 处理后的图片；若源图片不支持SubImage接口可能返回nil
// 逻辑: 先按短边居中裁剪为正方形，再缩放到目标尺寸
func squareImage(img image.Image, size int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()  // 原图宽度
	height := bounds.Dy() // 原图高度

	// 已是目标尺寸正方形直接返回，避免不必要处理
	if width == height && width == size {
		return img
	}

	// 按短边居中裁剪为正方形
	var cropped image.Image
	subImager, canCrop := img.(interface {
		SubImage(r image.Rectangle) image.Image
	})
	if width > height {
		// 宽图：从水平中间裁掉左右两边，裁剪区域宽=高=height
		if canCrop {
			x := (width - height) / 2
			cropped = subImager.SubImage(image.Rect(x, 0, x+height, height))
		}
	} else if height > width {
		// 高图：从垂直中间裁掉上下两边，裁剪区域宽=高=width
		if canCrop {
			y := (height - width) / 2
			cropped = subImager.SubImage(image.Rect(0, y, width, y+width))
		}
	} else {
		cropped = img // 已是正方形，无需裁剪
	}
	// 不支持 SubImage 的图片类型直接按原图缩放，避免类型断言 panic
	if cropped == nil {
		cropped = img
	}

	// 缩放到目标尺寸（Bicubic双三次插值，质量较好）
	return resize.Resize(uint(size), uint(size), cropped, resize.Bicubic)
}
