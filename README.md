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
gor, err := gorgeous.New(ctx) // 生成一个Gorgeous实例

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


