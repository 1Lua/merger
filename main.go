package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

const (
	DefaultSource = "dev"
	DefaultTarget = "master"
	DefaultTag    = ""
)

var mergeFlags = []cli.Flag{
	&cli.StringFlag{
		Name:        "source",
		DefaultText: DefaultSource,
		Aliases:     []string{"s"},
		Usage:       "Ветка для слияния из",
		Required:    false,
	},
	&cli.StringFlag{
		Name:        "target",
		DefaultText: DefaultTarget,
		Aliases:     []string{"t"},
		Usage:       "Ветка для слияния в",
		Required:    false,
	},
	&cli.StringFlag{
		Name:     "tag",
		Usage:    "Имя тега для коммита",
		Required: false,
	},
}

func mergeBranches(context *cli.Context, semVersionPart int) error {
	source := context.String("source")
	target := context.String("target")
	tag := context.String("tag")

	utility := GitUtility{repoPath: "."}

	if tag == "" {
		var err error
		tag, err = utility.getNewTag(target, semVersionPart)
		if err != nil {
			return fmt.Errorf("не удалось получить новый тег: %v", err)
		}
	}

	if err := utility.mergeBranches(source, target); err != nil {
		return fmt.Errorf("не удалось слить ветки: %v", err)
	}
	if err := utility.tagCommit(tag); err != nil {
		return fmt.Errorf("не удалось создать тег: %v", err)
	}
	fmt.Printf("Слияние и тегирование завершены с тегом %s.\n", tag)

	changes, err := utility.getChangeList(tag, "", "string")
	if err != nil {
		return fmt.Errorf("не удалось получить список изменений: %v", err)
	}
	fmt.Println("Изменения:")
	fmt.Println(changes)
	return nil
}

func main() {
	app := &cli.App{
		Name: "merger",
		Commands: []*cli.Command{
			&cli.Command{
				Name:  "major",
				Usage: "Слить ветки и увеличить старшую версию",
				Flags: mergeFlags,
				Action: func(context *cli.Context) error {
					return mergeBranches(context, SemVersionMajor)
				},
			},
			&cli.Command{
				Name:  "minor",
				Usage: "Слить ветки и увеличить минорную версию",
				Flags: mergeFlags,
				Action: func(context *cli.Context) error {
					return mergeBranches(context, SemVersionMinor)
				},
			},
			&cli.Command{
				Name:  "patch",
				Usage: "Слить ветки и увеличить патч-версию",
				Flags: mergeFlags,
				Action: func(context *cli.Context) error {
					return mergeBranches(context, SemVersionPatch)
				},
			},
			{
				Name:  "changes",
				Usage: "Получить список изменений для версии",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "version",
						Usage:    "Версия для получения списка изменений",
						Aliases:  []string{"v"},
						Required: true,
					},
					&cli.StringFlag{
						Name:        "format",
						Aliases:     []string{"f"},
						Usage:       "Формат вывода (text или json)",
						DefaultText: TextOutput,
					},
				},
				Action: func(c *cli.Context) error {
					version1 := c.String("version")
					format := c.String("format")

					utility := GitUtility{repoPath: "."}
					changes, err := utility.getChangeList(version1, "", format)
					if err != nil {
						return fmt.Errorf("не удалось получить список изменений: %v", err)
					}

					fmt.Println(changes)
					return nil
				},
			},
		},
		Authors: []*cli.Author{
			&cli.Author{
				Name:  "Matvey Sokolanov",
				Email: "msokolanov@gmail.com",
			},
		},
		Action: func(context *cli.Context) error {
			fmt.Printf("Запустите 'merger --help' для получения помощи.\n")
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
