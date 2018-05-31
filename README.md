# Gorgeous任务框架

## 简介
    Gorgeous是一个使用Go语言编写的任务框架, 可用于处理大量批处理任务。

## 如何使用Gorgeous编写一个后台任务

### 实现一个 ```provider.IFetchHandler```

``` go
// FetcherHandler的定义
type IFetchHandler interface {
	Name() string // fetcher的名字
	Size() int // 开启多少个goroutine同时运行
	Action() (interface{}, error) // 提供数据
	Interval() time.Duration // 每隔多少时间运行一次，如果是0，则表示不间断
	Close() error // 关闭相关资源
}


type fetcherDemo struct{}

func (f *fetcherDemo) Name() string {
	return "fetcher_demo"
}

func (f *fetcherDemo) Size() int {
	return 1
}

func (f *fetcherDemo) Action() (interface{}, error) {
	return "hello world", nil
}

func (f *fetcherDemo) Interval() time.Duration {
	return 1 * time.Second
}

func (f *fetcherDemo) Close() error {
	log.Println("fetcher is closed.")
	return nil
}

```


### 实现一个 ```provider.IWorkHandler```

```go
// WorkHandler
type IWorkHandler interface {
	Name() string // worker的名字
	Size() int // 同时开启多少个worker运行
	HandleData(interface{}) (interface{}, error) // 接收来fetcher里的提供的数据
	SetNext(IWorkHandler) // 可以将数据传递给下一个workHandler做处理,如果不需要传递给下级workHandler, 则空着方法即可
	Next() IWorkHandler // 返回下一个workHandler, 如果无下级workHandler，返回nil
	Close() error // 关闭资源
}


type workerDemo struct{}

func (h *workerDemo) Name() string {
	return "demo"
}
func (h *workerDemo) Size() int {
	return 1
}

func (h *workerDemo) HandleData(data interface{}) (interface{}, error) {
	log.Printf("got data: %+v", data)
	return "data that pass to next worker", fmt.Errorf("some error: data = %+v", data)
}

func (h *workerDemo) SetNext(provider.IWorkHandler) {
}

func (h *workerDemo) Next() provider.IWorkHandler {
	return nil
}

func (h *workerDemo) Close() error {
	log.Println("worker is closed")
	return nil
}
```    

### 开启任务

```go

ctx, cancel := context.WithCancel(context.Background()) // 启动一个context用于整体关闭任务
gor, err := NewGorgeous(ctx) // 生成一个Gorgeous实例

gor.Add("demo", &fetcherDemo{}, &workerDemo{}) //添加一个任务,提供名字，fetchHandler, workHandler

// 接受到Ctrl-C之后执行函数
var stop = func() {
    cancel()
    log.Printf("done")
}

// 开启任务
if err := gor.Start(stop); err != nil {
    log.Fatal("gorgeous start failed:", err)
}

```

![an example red dot](data:image/jpeg;base64,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B1rej1t40CbAcDFYnlyf3T+VHlyf3T+VAG9/b7/88xR/bznjyxWD5cn90/lSiOTI+U/lQB0Q0ZLgecXI384+tPTQ0Rw3mHg5rYt/9QmfQVPQAgGBiuW13/XJ9K6quX1xWaZdoJ4oA5+in+XJ/dP5UeXJ/dP5UAMop/lyf3T+VHlyf3T+VAHYaP8A8eY+tatZekAi0AIxzWpQAVTvLUXcXlE45q5RQBz39gp/z0NbkMflRrH12jFS0UAFFFFABRRRQB//0+kooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooA/9TpKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKACiiigAooooAKKKKAP/Z)
