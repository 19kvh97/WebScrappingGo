package factory

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/19kvh97/webscrappinggo/upworksdk/models"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type IEmployee interface {
	StartWorking()
}

type Result struct {
	models.IParcell
	Data string
}

type WorkFailed struct {
	ConfigID string
	Message  string
}

type Task func(context.Context) (*Result, error)

type StateMachine struct {
	Config models.Config
	Task   Task
	Result *Result
}

type Employee struct {
	IEmployee
	UpdateJobChan chan models.Config
	StopWork      chan struct{}
	ResultChan    chan models.IParcell
	ErrorChan     chan WorkFailed
}

func (e *Employee) StartWorking(initConfig models.Config) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("start-fullscreen", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),
	)
	allocCtx, _ := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))

	prepareStageChan := make(chan *StateMachine)
	doJobStageChan := make(chan *StateMachine)
	sendResultStageChan := make(chan *StateMachine)
	errStageChan := make(chan WorkFailed)

	go e.prepareJob(ctx, prepareStageChan, errStageChan, doJobStageChan)
	go e.doJob(ctx, doJobStageChan, errStageChan, sendResultStageChan)
	go e.SendResult(ctx, sendResultStageChan, errStageChan, doJobStageChan)
	go e.ErrorNotify(ctx, errStageChan)

	//init state
	prepareStageChan <- &StateMachine{
		Config: initConfig,
	}

	for {
		<-e.StopWork
		log.Println("Employee stop working")
		cancel()
		return
	}
}

func (e *Employee) prepareJob(ctx context.Context, input <-chan *StateMachine, errStageChan chan<- WorkFailed, nextStage chan<- *StateMachine) {
	log.Println("Prepare job")
	lastState := &StateMachine{}
	taskChan := make(chan chromedp.Tasks)
	for {
		select {
		case <-ctx.Done():
			return
		case stm := <-input:
			if stm == nil {
				errStageChan <- WorkFailed{
					ConfigID: "empty",
					Message:  "Config is empty",
				}
				continue
			}
			if !reflect.DeepEqual(lastState.Config, stm.Config) {
				if lastState != nil && stm != nil {
					log.Printf("Update config %s to new config %s", lastState.Config.Id, stm.Config.Id)
				}
				lastState.Config = stm.Config
				if lastState.Config.Mode == models.UNKNOWN {
					errStageChan <- WorkFailed{
						ConfigID: lastState.Config.Id,
						Message:  "Config mode undefined",
					}
					return
				}

				if lastState.Config.Interval <= 0 {
					errStageChan <- WorkFailed{
						ConfigID: lastState.Config.Id,
						Message:  fmt.Sprintf("Invalid interval : %d", lastState.Config.Interval),
					}
					return
				}

				go setUpChromeInstance(ctx, e.UpdateJobChan, taskChan)

				log.Println("set task")
				lastState.Task = func(ctx context.Context) (*Result, error) {
					var nodes []*cdp.Node
					var job models.Job

					log.Printf("link : %s", lastState.Config.Mode.GetLink())

					err := chromedp.Run(ctx,
						// navigate to site
						chromedp.Navigate(lastState.Config.Mode.GetLink()),
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
						Data: job.Title,
					}, nil
				}

			}

			nextStage <- lastState
		}
	}
}

func setUpChromeInstance(extCtx context.Context, configChan <-chan models.Config, taskChan <-chan chromedp.Tasks) {
	for {
		select {
		case <-extCtx.Done():
			return
		case cf := <-configChan:
			cookies := cf.Account.Cookie

			log.Printf("cookies length in PrepareTask %d", len(cookies))
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
				chromedp.Navigate(cf.Mode.GetLink()),
			}
			if err := chromedp.Run(extCtx, tasks); err != nil {
				log.Printf("err : %s", err.Error())
			}
		case tasks := <-taskChan:
			if err := chromedp.Run(extCtx, tasks); err != nil {
				log.Printf("err : %s", err.Error())
			}
		}
	}

}

func (e *Employee) doJob(ctx context.Context, input <-chan *StateMachine, errStageChan chan<- WorkFailed, nextStage chan<- *StateMachine) {
	lastState := &StateMachine{}
	for {
		select {
		case <-ctx.Done():
			return
		case stm := <-input:
			if stm == nil {
				continue
			}
			if !reflect.DeepEqual(lastState.Config, stm.Config) {
				lastState.Result = nil
				nextStage <- lastState
				continue
			}
			lastState = stm
			if stm.Task == nil {
				errStageChan <- WorkFailed{
					ConfigID: stm.Config.Id,
					Message:  "Task is nil",
				}
			} else {
				log.Println("on do job")
				result, err := stm.Task(ctx)
				if err != nil {
					errStageChan <- WorkFailed{
						ConfigID: stm.Config.Id,
						Message:  fmt.Sprintf("Do job error : %s", err.Error()),
					}
				}
				lastState.Result = result
				nextStage <- lastState
			}
		case cf := <-e.UpdateJobChan:
			if !reflect.DeepEqual(lastState.Config, cf) {
				lastState.Config = cf
			}
		}
	}
}

func (e *Employee) SendResult(ctx context.Context, input <-chan *StateMachine, errStageChan chan<- WorkFailed, nextStage chan<- *StateMachine) {
	for {
		select {
		case <-ctx.Done():
			return
		case stm := <-input:
			if stm == nil {
				continue
			}
			if stm.Result != nil {
				e.ResultChan <- stm.Result
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Duration(stm.Config.Interval) * time.Millisecond):
					stm.Result = nil
					nextStage <- stm
				}
			} else {
				nextStage <- stm
			}
		}
	}
}

func (e *Employee) ErrorNotify(ctx context.Context, input <-chan WorkFailed) {
	for {
		select {
		case <-ctx.Done():
			return
		case wf := <-input:
			e.ErrorChan <- wf
			close(e.StopWork)
		}
	}
}
