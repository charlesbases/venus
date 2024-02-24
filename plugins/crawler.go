package plugins

import (
	"context"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/charlesbases/lifecycle"
	"github.com/charlesbases/progressbar"
	"github.com/charlesbases/salmon"
	"github.com/urfave/cli"

	"github.com/charlesbases/venus/logger"
	"github.com/charlesbases/venus/sender"
	"github.com/charlesbases/venus/types"
	"github.com/charlesbases/venus/website"
)

// crawler .
type crawler struct {
	concurrent int

	recorder recorder
	producer producer

	lc lifecycle.Lifecycle
}

// Crawler .
func Crawler(ctx *cli.Context) *crawler {
	return new(crawler).prepare(ctx)
}

// prepare .
func (cr *crawler) prepare(ctx *cli.Context) *crawler {
	cr.lc = lifecycle.New()
	cr.lc.Append(
		// logger
		lifecycle.Hook{
			OnStart: func(ctx context.Context) error {
				return logger.Create()
			},
			OnStop: func(ctx context.Context) error {
				logger.Close()
				return nil
			},
		},
		// clerk
		lifecycle.Hook{
			OnStart: func(ctx context.Context) error {
				r, err := newRecorder()
				cr.recorder = r
				return err
			},
			OnStop: func(ctx context.Context) error {
				return cr.recorder.Close()
			},
		},
	)

	// 并发数
	cr.concurrent = int(ctx.Uint("c"))

	// producer
	cr.producer = fileProducer(ctx.String("f"))
	if len(ctx.Args()) != 0 {
		cr.producer = argsProducer(ctx.Args())
	}
	return cr
}

// Start .
func (cr *crawler) Start() error {
	if err := cr.lc.Start(context.Background()); err != nil {
		return err
	}
	return cr.run()
}

// Stop .
func (cr *crawler) Stop() error {
	return cr.lc.Stop(context.Background())
}

// run .
func (cr *crawler) run() error {
	hooks, err := cr.producer.Produce()
	if err != nil {
		return err
	}

	bar := progressbar.NewReader()

	// 并发下载
	pool, err := salmon.NewPool(cr.concurrent)
	if err != nil {
		return err
	}

	var notify = make(chan os.Signal)
	go func() {
		signal.Notify(notify, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
		for {
			select {
			case <-notify:
				pool.Stop()
				return
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(len(hooks))

	for i := range hooks {
		go func(hook website.WebHook) {
			defer wg.Done()

			// 获取用户信息
			user, err := hook.UserInfor()
			if err != nil {
				logger.Error(err)
				return
			}

			if len(user.Videos) == 0 {
				return
			}

			// 视频文件夹
			root := types.Storage.PathJoin(user.Root, user.ID).MkdirAll()

			pb := bar.NewProgress(user.Title(), uint(len(user.Videos)))

			for j := range user.Videos {
				header := user.Videos[j]
				pool.Invoke(
					func(cancel func()) {
						defer pb.Incr(1)

						// 查看当前视频是否下载
						if !root.FileIsExists(header.VideoID) {
							// 解析视频
							video, err := hook.ParseHeader(header)
							if err != nil {
								logger.Error(err)
								return
							}

							if len(video.Parts) == 0 {
								return
							}

							// 下载并写入文件
							if err := root.CreateFile(video.VideoID).Write(
								func(w io.Writer) error {
									for ii := range video.Parts {
										if err := hook.Sender().SendRequest(
											video.Parts[ii].String(),
											sender.WriteTo(w),
											video.Metadata...,
										); err != nil {
											return err
										}
									}
									return nil
								},
							); err != nil {
								logger.Error(err)
								return
							}
						}

						// record
						cr.recorder.Write(
							user.Root, header.VideoID, recordUser(user.ID),
							recordLink(header.LinkPath.String()), recordTitle(header.Title),
						)
					},
				)
			}
		}(hooks[i])
	}

	wg.Wait()
	pool.Wait()
	bar.Close()

	close(notify)
	return nil
}
