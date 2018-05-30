package supervisor

func Supervisor(worker func(chan struct{}), size int) {
	restart := make(chan struct{})
	go worker(restart)
	restartCount := 1

	go func() {
		for {
			select {
			case <-restart:
				restartCount++
				if restartCount > size {
					return
				}

				go worker(restart)
			}
		}
	}()
}

func Worker(f func()) func(chan struct{}) {
	return func(restart chan struct{}) {
		defer func() {
			if err := recover(); err != nil {
				restart <- struct{}{}
			}
		}()

		f()
	}
}
