package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"log"
	"os"
	"sort"
)

var logger *zap.Logger

func init() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	logger.Info("logger start server")
}


func main() {

	//arguments()
	//Flags()
	//DestinationFlags()
	//PleaceholderValue()
	//ShortName()

	Subcommands()
}


//级别优先级，如果有子命令,那么主命令的Flage会不起作用，
//也即是说有子命令，输入了子命令，上级命令的参数都会失效

//Args 和 --参数 -参数 只能有一种形式，要么是 common + 参数1 参数2  要么是common + --参数名 参数只  --参数名  参数值（错误理解，可同时存在）
func Subcommands() {

	//Action内部可以访问有值或者在app.Run之后也会有值，在没有run之前不可能有值
	var dest int
	//var dest3 *int
	app := &cli.App{
		UseShortOptionHandling:true,
		EnableBashCompletion:true,
		Action: func(cCtx *cli.Context) error {
			//fmt.Println(cCtx.String("port"))
			fmt.Printf("hello %v \n",cCtx.String("port"))
			//fmt.Println(cCtx.Args().Len())

			//Action内部可以访问有值
			fmt.Println("======dest1=====",dest)
			//dest3 = &dest
			return nil
		},
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "port",
				Usage:       "Use a randomized port",
				Value:       0,
				Aliases: []string{"p"},
				Destination: &dest,
			},
		},
		Commands: []*cli.Command{
			{
				Name: "add",
				Aliases: []string{"a"},
				Usage: "add a task to the list",
				Category: "template",//子命令类别
				Action: func(cCtx *cli.Context) error {
					fmt.Println("added task: ",cCtx.Args().First())
					fmt.Println("added task host: ",cCtx.String("o"))
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "options",
						Usage:       "Use a default localhost host",
						Value:       "localhost",
						Aliases: []string{"o"},
					},
				},
			},{
				Name: "complete",
				Aliases: []string{"c"},
				Category: "complete",
				Usage: "complete a task to the list",
				Action: func(cCtx *cli.Context) error {
					fmt.Println("complete task: ",cCtx.Args().First())
					return nil
				},
			},{
				Name: "template",
				Aliases: []string{"t"},
				Category: "template",//子命令类别
				Usage: "options for task templates",
				Subcommands: []*cli.Command{
					{
						Name:  "add",
						Usage: "add a new template",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "opt",
								Usage:       "Use a default localhost host",
								Value:       "opts",
								Aliases: []string{"oo"},
							},
						},
						Action: func(cCtx *cli.Context) error {
							fmt.Println("new task template: ", cCtx.Args().First())
							fmt.Println("new task template: ", cCtx.String("oo"))
							return nil
						},
					},
					{
						Name:  "remove",
						Usage: "remove an existing template",
						Action: func(cCtx *cli.Context) error {
							fmt.Println("removed task template: ", cCtx.Args().First())
							return nil
						},
					},
				},
			},
		},
	}

	//只能在Action内部使用 （错误理解，这里还没有Run肯定取不到命令行输入信息）
	fmt.Println("======dest2=====",dest)
	//fmt.Println("======dest3=====",dest3)

	if err := app.Run(os.Args); err != nil {
		logger.Fatal("subcommands error",zap.Error(err))
	}
}



func DefaultText() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:        "port",
				DefaultText: "random",
				Usage:       "Use a randomized port",
				Value:       0,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal("default text error",zap.Error(err))
	}
}


func valueFromEnv() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "lang",
				Usage:       "language for the greeting",
				Aliases:     []string{"l"},
				EnvVars:     []string{"GO_PATH"},
			},
		},
		//Action: func(cCtx *cli.Context) error {
		//
		//	fmt.Printf("%q \n",cCtx.String("l"))
		//
		//	return nil
		//},
	}


	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func SortFlagAndComm() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "lang",
				Aliases: []string{"l","ll"},
				Value: "english",
				Usage: "Language for the greeting",
			},
			&cli.StringFlag{
				Name: "config",
				Aliases: []string{"c","cc"},
				Usage: "Load configuration from `FILE`",
			},
		},
		Commands: []*cli.Command{
			{
				Name:"complete",
				Aliases:  []string{"c"},
				Usage: "complete a task on the list",
				Action: func(cCtx *cli.Context) error {
					fmt.Printf("%q \n",cCtx.String("c"))
					return nil
				},
			},
			{
				Name:"add",
				Aliases:  []string{"a"},
				Usage: "add a task to the list",
				Action: func(cCtx *cli.Context) error {
					fmt.Printf("%q \n",cCtx.String("a"))
					return nil
				},
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}


func ShortName() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "lang",
				Aliases: []string{"l"},
				Value: "english",
				Usage: "language for the greeting",
			},
		},
		Action: func(cCtx *cli.Context) error {
			fmt.Printf("hello %v \n",cCtx.String("lang"))
			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func PleaceholderValue() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "config",
				Aliases:  []string{"c"},
				Usage: "Load configuration from `FILE`",
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func DestinationFlags() {
	var language string

	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "lang",
				Value: "english",
				Usage: "language for the greeting",
				Destination: &language,
			},
		},
		Action: func(cCtx *cli.Context) error {
			name := "someone"
			if cCtx.NArg() > 0 {
				name = cCtx.Args().Get(0)
			}
			if language == cCtx.String("lang") && language == "spanish" {
				fmt.Println("Hola", name)
			} else {
				fmt.Println("Hello",name)
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil{
		logger.Fatal("create ursafe error ",zap.Error(err))
	}

}




func Flags() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "lang",
				Value: "english",
				Usage: "language for the greeting",
			},
		},
		Action: func(cCtx *cli.Context) error {
			name := "Nefertiti"
			if cCtx.NArg() > 0 {
				name = fmt.Sprintf("%v",cCtx.Args().Get(0))
			}
			//fmt.Printf("Hello %q",cCtx.Args().Get(0))
			if cCtx.String("lang") == "spanish" {
				fmt.Println("Hola",name)
			}else {
				fmt.Println("hello",name)
			}
			return nil
		},
	}

	if err :=  app.Run(os.Args); err != nil {
		logger.Fatal("create ursafe error ",zap.Error(err))
	}

}


func arguments() {
	app := &cli.App{
		Action: func(cCtx *cli.Context) error {
			fmt.Printf("Hello %q",cCtx.Args().Get(0))
			return nil
		},
	}

	err := app.Run(os.Args)

	if  err != nil {
		logger.Fatal("create ursafe error ",zap.Error(err))
	}
}


func simpleExam() {


	//App is the main structure of a cli application. //App是cli应用的主要结构。

	//It is recommended that an app be created with the cli.NewApp() function

	//建议使用cli.NewApp()函数创建应用程序

	///err := (&cli.App{}).Run(os.Args)//This app will run and show help text, but is not very useful. 没什么用，只是展示一些帮助信息而已

	/**
	$ ./ursafe.exe
	NAME:
	   ursafe.exe - A new cli application

	USAGE:
	   ursafe.exe [global options] command [command options] [arguments...]

	COMMANDS:
	   help, h  Shows a list of commands or help for one command

	GLOBAL OPTIONS:
	   --help, -h  show help (default: false)
	*/
	//if err != nil {
	//	panic(err)
	//}

	app := &cli.App{
		Name:                   "greet",
		//HelpName:               "make an explosive entrance",
		Usage:                  "fight the loneliness!",
		//UsageText:              "",
		//ArgsUsage:              "",
		//Version:                "",
		//Description:            "",
		//Commands:               nil,
		//Flags:                  nil,
		//EnableBashCompletion:   false,
		//HideHelp:               false,
		//HideHelpCommand:        false,
		//HideVersion:            false,
		//BashComplete:           nil,
		//Before:                 nil,
		//After:                  nil,
		Action: func(ctx *cli.Context) error {
			//fmt.Println("boom! I say!")
			fmt.Println("Hello friend!")
			return nil
		},
		//CommandNotFound:        nil,
		//OnUsageError:           nil,
		//Compiled:               time.Time{},
		//Authors:                nil,
		//Copyright:              "",
		//Reader:                 nil,
		//Writer:                 nil,
		//ErrWriter:              nil,
		//ExitErrHandler:         nil,
		//Metadata:               nil,
		//ExtraInfo:              nil,
		//CustomAppHelpTemplate:  "",
		//UseShortOptionHandling: false,
		//Suggest:                false,
	}

	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	if err := app.Run(os.Args); err != nil {
		logger.Fatal("create ursafe error ",zap.Error(err))
	}
}