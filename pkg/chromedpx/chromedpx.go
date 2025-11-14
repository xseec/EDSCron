package chromedpx

import (
	"context"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/cdproto/browser"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

// DPUrl 定义网页URL和点击操作配置
type DPUrl struct {
	Url    string   `json:"url" yaml:"url"`       // 目标网页URL
	Clicks []string `json:"clicks" yaml:"clicks"` // 点击目标列表，可以是innerText、id、class或style
	// 以*开头的点击项为可选操作(如弹窗确认按钮)
}

// DPOuter 定义结果提取配置
type DPOuter struct {
	OnlyText bool   `json:"only_text" yaml:"only_text"` // 纯文本提取
	Selector string `json:"selector" yaml:"selector"`   // 结果选择器，空表示新标签页URL
	Pattern  string `json:"pattern" yaml:"pattern"`     // 结果匹配正则表达式
	Host     string `json:"host" yaml:"host"`           // 域名补全，用于子域名结果
}

// DP 定义完整的浏览器自动化配置
type DP struct {
	IsVisible   bool    `json:"is_visible" yaml:"is_visible"`     // 是否显示浏览器窗口
	DownloadDir string  `json:"download_dir" yaml:"download_dir"` // 下载目录
	Urls        []DPUrl `json:"urls" yaml:"urls"`                 // 网页操作流程
	Outer       DPOuter `json:"outer" yaml:"outer"`               // 结果提取配置
}

var (
	optionalSelReg = regexp.MustCompile(`^\*(\S+)`)
	regexpSelReg   = regexp.MustCompile(`^\+(\S+)-(\S+)`)
	pdfUrlReg      = regexp.MustCompile(`(?i)^https?://.*(?:\.pdf$|documentType=pdf)`)
)

// Run 执行浏览器自动化流程
func (dp *DP) Run(ctx context.Context, content *string) error {
	if len(dp.Urls) == 0 {
		return fmt.Errorf("URL配置不能为空")
	}

	if content == nil {
		return fmt.Errorf("结果存储指针不能为nil")
	}

	// 初始化浏览器上下文
	cctx, cancels := InitDP(ctx, dp.IsVisible)
	defer func() {
		for _, c := range cancels {
			c()
		}
	}()

	// 指定下载目录（单次生效），处理点击元素直接下载而无法捕获的问题
	if len(dp.DownloadDir) > 0 {
		absPath, err := filepath.Abs(dp.DownloadDir)
		if err != nil {
			return fmt.Errorf("获取下载目录绝对路径失败: %w", err)
		}
		err = chromedp.Run(cctx,
			browser.SetDownloadBehavior(browser.SetDownloadBehaviorBehaviorAllow).
				WithDownloadPath(absPath).
				WithEventsEnabled(true),
		)
		if err != nil {
			return fmt.Errorf("设置下载行为失败: %w", err)
		}
	}

	// 执行每个URL的操作流程
	for i, u := range dp.Urls {
		if err := dp.processURL(cctx, u, i); err != nil {
			return fmt.Errorf("处理URL[%s]失败: %w", u.Url, err)
		}
	}

	// 提取最终结果
	*content = html.UnescapeString(useOuter(cctx, dp.Outer))
	if *content == "" {
		return fmt.Errorf("未获取到有效结果，配置: %+v", *dp)
	}

	return nil
}

// processURL 处理单个URL的操作流程
func (dp *DP) processURL(ctx context.Context, u DPUrl, urlIndex int) error {
	// 导航到目标URL
	// 重试配置
	maxRetries := 5               // 最大重试次数
	baseDelay := 10 * time.Second // 基础延迟
	var lastErr error
	// 重试导航
	for i := range maxRetries {
		// 指数退避延迟（第一次立即重试）
		if i > 0 {
			delay := baseDelay * time.Duration(1<<(i-1)) // 10s, 20s, 40s...
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// 尝试导航
		err := chromedp.Run(ctx, chromedp.Navigate(u.Url))
		if err == nil {
			break // 成功则退出重试
		}
		lastErr = fmt.Errorf("导航重试(%d/%d)失败: %w", i+1, maxRetries, err)
	}

	if lastErr != nil {
		return lastErr
	}

	// 执行每个点击操作
	for j, click := range u.Clicks {
		if err := dp.processClick(ctx, click, urlIndex, j, len(u.Clicks)-1); err != nil {
			return fmt.Errorf("点击操作[%s]失败: %w", click, err)
		}
	}

	return nil
}

// processClick 处理单个点击操作
func (dp *DP) processClick(ctx context.Context, click string, urlIndex, clickIndex, lastClickIndex int) error {

	// 处理可选点击项，如"*确认"：随机出现的弹窗确认按钮
	optional := false
	if subs := optionalSelReg.FindStringSubmatch(click); len(subs) == 2 {
		click = subs[1]
		optional = true
	}

	// 处理正则表达式点击项，如"+代理购电-2025年0?8月"：从"代理购电"节点列表中匹配"2025年0?8月"的节点
	detailClick := ""
	if subs := regexpSelReg.FindStringSubmatch(click); len(subs) == 3 {
		click = subs[1]
		detailClick = subs[2]
	}

	// 查找目标节点
	var nodes []*cdp.Node
	searchSecond := time.Second * 20
	if optional {
		searchSecond = time.Second * 10
	}
	searchCtx, cancel := context.WithTimeout(ctx, searchSecond)
	defer cancel()

	if err := chromedp.Run(searchCtx, chromedp.WaitReady(click), chromedp.Nodes(click, &nodes, chromedp.BySearch)); !optional && err != nil {
		return fmt.Errorf("查找节点失败: %w", err)
	}

	// 处理未找到节点的情况
	if len(nodes) == 0 {
		if optional {
			return nil // 可选操作允许节点不存在
		}
		return fmt.Errorf("未找到匹配节点: %s", click)
	}

	// 生成选择器并执行点击
	node := nodes[0]
	if detailClick != "" {
		for _, n := range nodes {
			if regexp.MustCompile(detailClick).MatchString(n.NodeValue) {
				node = n
				break
			}
		}
	}

	selector := useSelectorOf(node)
	return dp.executeClick(ctx, selector, urlIndex, clickIndex, lastClickIndex)
}

// executeClick 执行实际的点击操作
func (dp *DP) executeClick(ctx context.Context, selector string, urlIndex, clickIndex, lastClickIndex int) error {
	// 最后一个点击操作不需要等待
	duration := 1
	if urlIndex == len(dp.Urls)-1 && clickIndex == lastClickIndex {
		duration = 0
	}

	actions := []chromedp.Action{
		chromedp.WaitVisible(selector),
		chromedp.RemoveAttribute(selector, "target"), // 防止新标签页打开
		chromedp.Sleep(5 * time.Second),              // 确保页面加载完成
		chromedp.Click(selector),
	}

	if duration > 0 {
		actions = append(actions, chromedp.Sleep(time.Duration(duration)*time.Second))
	}

	return chromedp.Run(ctx, actions...)
}

// InitDP 初始化浏览器上下文
func InitDP(ctx context.Context, isVisible bool) (context.Context, []context.CancelFunc) {
	// 基础配置
	ops := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", !isVisible),
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0.0.0 Safari/537.36 Edg/128.0.0.0`),
	}

	// Docker环境特殊配置
	if os.Getenv("RUNTIME") == "docker" {
		ops = append(ops,
			chromedp.NoSandbox,
			chromedp.DisableGPU,
			chromedp.Flag("disable-dev-shm-usage", true),
			chromedp.Flag("remote-debugging-address", "0.0.0.0"),
			chromedp.Flag("remote-debugging-port", "9222"),
		)
	}

	// 合并默认配置
	ops = append(chromedp.DefaultExecAllocatorOptions[:], ops...)

	// 创建上下文链
	ctx, c0 := context.WithTimeout(ctx, 5*time.Minute)
	ctx, c1 := chromedp.NewExecAllocator(ctx, ops...)
	ctx, c2 := chromedp.NewContext(ctx)
	cancels := []context.CancelFunc{c0, c1, c2}

	// 隐藏自动化特征
	if err := chromedp.Run(ctx, hideWebDriverFlag()); err != nil {
		// 非致命错误，仅记录
		fmt.Printf("警告: 隐藏webdriver标志失败: %v\n", err)
	}

	return ctx, cancels
}

// hideWebDriverFlag 隐藏浏览器自动化特征
func hideWebDriverFlag() chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		_, err := page.AddScriptToEvaluateOnNewDocument(
			"Object.defineProperty(navigator, 'webdriver', { get: () => undefined });",
		).Do(ctx)
		return err
	})
}

// useOuter 提取操作结果
func useOuter(ctx context.Context, o DPOuter) string {
	if url := getExtraTabUrl(ctx); url != "" {
		return url
	}

	searchCtx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()
	// 提取当前页面内容
	var result string
	// 纯文本提取
	if o.OnlyText {
		if err := chromedp.Run(searchCtx, chromedp.Text(o.Selector, &result)); err != nil {
			return ""
		}
		return result
	}

	// HTML提取
	if err := chromedp.Run(searchCtx, chromedp.OuterHTML(o.Selector, &result)); err != nil {
		return ""
	}

	// 应用正则匹配
	if o.Pattern != "" {
		reg := regexp.MustCompile(o.Pattern)
		subs := reg.FindStringSubmatch(result)
		if len(subs) > 1 {
			result = subs[1]
		} else if len(subs) == 1 {
			result = subs[0]
		}
	}

	// 补全域名
	if o.Host != "" && !strings.HasPrefix(result, o.Host) {
		return fmt.Sprintf("%s/%s", strings.TrimRight(o.Host, "/"), strings.TrimLeft(result, "/"))
	}

	return result
}

// getExtraTabUrl 获取新标签页URL
func getExtraTabUrl(ctx context.Context) string {
	// 时机一：最后一个click打开标签页
	ch := chromedp.WaitNewTarget(ctx, func(info *target.Info) bool {
		return info.URL != ""
	})

	select {
	case <-time.After(3 * time.Second):
		break
	case id := <-ch:
		nCtx, _ := chromedp.NewContext(ctx, chromedp.WithTargetID(id))
		var url string
		if err := chromedp.Run(nCtx, chromedp.Location(&url)); err != nil {
			return ""
		}
		return url
	}

	// 时机二：过程中click打开标签页，但是后续click可选，额外浪费了时间不能用"WaitNewTarget"
	targets, _ := chromedp.Targets(ctx)
	for _, t := range targets {
		if pdfUrlReg.MatchString(t.URL) {
			return t.URL
		}
	}

	return ""
}

// useSelectorOf 生成最优选择器
func useSelectorOf(n *cdp.Node) string {
	// 无效标签向上查找父节点
	invalidTags := map[string]bool{"": true, "text": true, "span": true}
	if invalidTags[n.LocalName] && n.Parent != nil {
		return useSelectorOf(n.Parent)
	}

	// 优先使用带ID的XPath
	if id := n.AttributeValue("id"); id != "" {
		return fmt.Sprintf("//*[@id='%s']", id)
	}

	// 回退到完整XPath
	return n.FullXPathByID()
}
