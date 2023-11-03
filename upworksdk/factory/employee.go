package factory

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/19kvh97/webscrappinggo/upworksdk/models"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type Result struct {
	models.IParcell
	Data string
}

type LoopState int

const (
	CREATE_ENV_STATE LoopState = iota
	CREATE_JOB_STATE
	SLEEP_STATE
)

type ErrorType int

const (
	LOG ErrorType = iota
	WARNING
	CRITICAL
)

type ErrorMessage struct {
	ConfigID string
	Message  string
	Type     ErrorType
}

type Task func(context.Context) (*Result, error)

type Employee struct {
	updateJobChan chan models.Config
	loopStateChan chan LoopState
	StopWork      chan struct{}
	ResultChan    chan models.IParcell
	ErrorChan     chan ErrorMessage
	State         LoopState
}

func (e *Employee) StartWorking(initConfig models.Config) {
	e.updateJobChan = make(chan models.Config)
	e.loopStateChan = make(chan LoopState)

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("start-fullscreen", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),
	)
	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	taskChan := make(chan Task)
	errChan := make(chan ErrorMessage)
	resultChan := make(chan models.IParcell)

	go setUpChromeInstance(ctx, taskChan, errChan, resultChan)

	//init state
	lastCf := initConfig

	go func() {
		time.Sleep(time.Second)
		e.UpdateConfig(initConfig)
	}()

	isUpdateConfig := false

	for {
		select {
		case <-e.StopWork:
			log.Println("Employee stop working")
			cancel()
			return
		case job := <-e.updateJobChan:
			log.Println("Updatejobchan")
			lastCf = job
			if e.State == SLEEP_STATE {
				isUpdateConfig = true
			} else {
				e.updateLoopState(CREATE_ENV_STATE)
			}
		case e.State = <-e.loopStateChan:
			log.Println("newLoop")
			switch e.State {
			case CREATE_ENV_STATE:
				task, err := e.createEnv(lastCf)
				if err != nil {
					errChan <- ErrorMessage{
						ConfigID: lastCf.Id,
						Message:  err.Error(),
						Type:     CRITICAL,
					}
					continue
				}
				taskChan <- task
				isUpdateConfig = false // update done and reset
			case CREATE_JOB_STATE:
				taskChan <- e.createJob(lastCf)
			case SLEEP_STATE:
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Duration(lastCf.Interval) * time.Millisecond):
					if isUpdateConfig {
						e.updateLoopState(CREATE_ENV_STATE)
					} else {
						e.updateLoopState(CREATE_JOB_STATE)
					}
				}
			}
		case msg := <-errChan:
			log.Println("onErr")
			e.ErrorChan <- msg
			switch msg.Type {
			case LOG:
				log.Printf("log with config %s : %s", msg.ConfigID, msg.Message)
				e.updateLoopState(CREATE_JOB_STATE)
			case WARNING:
				log.Printf("warning with config %s : %s", msg.ConfigID, msg.Message)
				e.updateLoopState(CREATE_JOB_STATE)
			case CRITICAL:
				close(e.StopWork)
			}
		case rs := <-resultChan:
			log.Println("onResult")
			if res, ok := rs.(*Result); ok && res != nil {
				e.ResultChan <- res
				e.updateLoopState(SLEEP_STATE)
			} else {
				e.updateLoopState(CREATE_JOB_STATE)
			}

		}

	}
}

func (e *Employee) updateLoopState(state LoopState) {
	go func() {
		e.loopStateChan <- state
	}()
}

func (e *Employee) UpdateConfig(cf models.Config) {
	go func() {
		e.updateJobChan <- cf
	}()
}

func (e *Employee) createEnv(cf models.Config) (Task, error) {
	if cf.Mode == models.UNKNOWN {
		return nil, errors.New("invalid RunningMode")
	}
	if cf.Interval <= 0 {
		return nil, fmt.Errorf("invalid interval %d", cf.Interval)
	}

	return func(ctx context.Context) (*Result, error) {
		cookies := cf.Account.Cookie

		runningmode := cf.Mode

		log.Printf("cookies length in PrepareTask %d", len(cookies))
		var nodes []*cdp.Node
		tasks := chromedp.Tasks{
			chromedp.ActionFunc(func(ctx context.Context) error {
				// create cookie expiration
				expr := cdp.TimeSinceEpoch(time.Now().Add(180 * 24 * time.Hour))
				// add cookies to chrome
				for _, cookie := range cookies {
					err := network.SetCookie(cookie.Name, cookie.Value).
						WithExpires(&expr).
						WithDomain("www.upwork.com").
						WithHTTPOnly(false).
						Do(ctx)
					if err != nil {
						return err
					}
				}
				return nil
			}),

			// navigate to site
			chromedp.Navigate(runningmode.GetLink()),
			chromedp.Nodes("section.up-card-section.up-card-list-section.up-card-hover", &nodes, chromedp.ByQueryAll, chromedp.AtLeast(0)),
		}
		if err := chromedp.Run(ctx, tasks); err != nil {
			return nil, err
		}
		if len(nodes) == 0 {
			return nil, errors.New("err : Can't find any node")
		}
		return nil, nil
	}, nil
}

func (e *Employee) createJob(cf models.Config) Task {
	log.Println("createJob")
	return func(ctx context.Context) (*Result, error) {
		var nodes []*cdp.Node
		var job models.Job
		err := chromedp.Run(ctx,
			chromedp.Navigate(cf.Mode.GetLink()),
			chromedp.Nodes("section.up-card-section.up-card-list-section.up-card-hover", &nodes, chromedp.ByQueryAll, chromedp.AtLeast(0)),
			chromedp.Sleep(3*time.Second))

		if err != nil {
			return nil, err
		}
		if len(nodes) == 0 {
			return nil, errors.New("err : Can't find any node")
		}

		for _, node := range nodes {
			err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
				res, err := dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
				if err != nil {
					return err
				}
				doc, err := goquery.NewDocumentFromReader(strings.NewReader(res))
				if err != nil {
					return err
				}

				doc.Find("section.up-card-section.up-card-list-section.up-card-hover").Each(func(index int, info *goquery.Selection) {
					job.ImportData(info)
				})
				return nil
			}))
			if err != nil {
				return nil, err
			}
		}
		return &Result{
			Data: "HelloWorld",
		}, nil
	}
}

func setUpChromeInstance(extCtx context.Context, taskChan <-chan Task, errChan chan<- ErrorMessage, resultChan chan<- models.IParcell) {
	for {
		select {
		case <-extCtx.Done():
			return
		case tasks := <-taskChan:
			log.Println("Receive new tasks")
			result, err := tasks(extCtx)
			if err != nil {
				errChan <- ErrorMessage{
					Type:    WARNING,
					Message: err.Error(),
				}
			} else {
				resultChan <- result
			}
		}
	}

}
