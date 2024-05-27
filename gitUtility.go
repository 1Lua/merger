package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const (
	SemVersionMajor = iota
	SemVersionMinor
	SemVersionPatch
)

const (
	TextOutput = "text"
	JsonOutput = "json"
)

type Commit struct {
	Hash    string `xml:"hash" json:"hash"`
	Author  string `xml:"author" json:"author"`
	Email   string `xml:"email" json:"email"`
	Date    string `xml:"date" json:"date"`
	Message string `xml:"message" json:"message"`
}

type Commits struct {
	Commits []Commit `xml:"commit"`
}

type GitUtility struct {
	repoPath string
}

func (g *GitUtility) mergeBranches(sourceBranch, targetBranch string) error {
	tempBranch := "temp-merge-branch"

	log.Printf("Переключение на ветку %s", targetBranch)
	cmd := exec.Command("git", "checkout", targetBranch)
	cmd.Dir = g.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Ошибка при переключении на ветку %s: %s\n%s", targetBranch, err, string(output))
		return err
	}

	log.Printf("Создание временной ветки %s", tempBranch)
	cmd = exec.Command("git", "checkout", "-b", tempBranch)
	cmd.Dir = g.repoPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("Ошибка при создании временной ветки %s: %s\n%s", tempBranch, err, string(output))
		return err
	}

	log.Printf("Слияние ветки %s во временную ветку %s", sourceBranch, tempBranch)
	cmd = exec.Command("git", "merge", sourceBranch)
	cmd.Dir = g.repoPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("Ошибка при слиянии ветки %s во временную ветку %s: %s\n%s", sourceBranch, tempBranch, err, string(output))

		log.Printf("Отмена слияния и удаление временной ветки %s", tempBranch)
		cmd = exec.Command("git", "merge", "--abort")
		cmd.Dir = g.repoPath
		cmd.CombinedOutput()

		cmd = exec.Command("git", "checkout", targetBranch)
		cmd.Dir = g.repoPath
		cmd.CombinedOutput()

		cmd = exec.Command("git", "branch", "-D", tempBranch)
		cmd.Dir = g.repoPath
		cmd.CombinedOutput()

		return err
	}

	log.Printf("Переключение на ветку %s", targetBranch)
	cmd = exec.Command("git", "checkout", targetBranch)
	cmd.Dir = g.repoPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("Ошибка при переключении на ветку %s: %s\n%s", targetBranch, err, string(output))
		return err
	}

	cmd = exec.Command("git", "merge", sourceBranch)
	cmd.Dir = g.repoPath
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("Ошибка при слиянии временной ветки %s в %s: %s\n%s", sourceBranch, targetBranch, err, string(output))
		return err
	}

	log.Printf("Удаление временной ветки %s", tempBranch)
	cmd = exec.Command("git", "branch", "-D", tempBranch)
	cmd.Dir = g.repoPath
	cmd.CombinedOutput()

	return nil
}

func (g *GitUtility) tagCommit(tagName string) error {
	log.Printf("Проставление тега %s", tagName)
	cmd := exec.Command("git", "tag", tagName)
	cmd.Dir = g.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Ошибка при проставлении тега %s: %s\n%s", tagName, err, string(output))
		return err
	}
	return nil
}

func (g *GitUtility) getTagList(branchName string) ([]string, error) {
	log.Printf("Получение списка тегов для ветки %s", branchName)
	cmd := exec.Command("git", "tag", "--merged", branchName)
	cmd.Dir = g.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Ошибка при получении списка тегов: %s\n%s", err, string(output))
		return nil, err
	}
	tags := strings.Split(string(output), "\n")
	var validTags []string
	// Регулярное выражение для поиска тегов вида x.y.z
	tagRegex := regexp.MustCompile(`^v?(\d+\.\d+\.\d+)$`)
	for _, tag := range tags {
		if tagRegex.MatchString(tag) {
			validTags = append(validTags, tag)
		}
	}
	if len(validTags) > 0 {
		sort.Slice(validTags, func(i, j int) bool {
			return compareTags(validTags[i], validTags[j]) < 0
		})
	}
	return validTags, nil
}

func (g *GitUtility) getNewTag(branchName string, semVersionPart int) (string, error) {
	var validTags, err = g.getTagList(branchName)
	if err != nil {
		return "", err
	}
	if len(validTags) == 0 {
		return "v1.0.0", nil // Если нет тегов, начинаем с 1.0.0
	}

	latestTag := validTags[len(validTags)-1]
	return incrementTag(latestTag, semVersionPart), nil
}

func (g *GitUtility) getChangeList(version1, version2, format string) (string, error) {
	if version2 == "" {
		tags, err := g.getTagList("HEAD")
		if err != nil {
			return "", err
		}
		for i := len(tags) - 1; i >= 0; i-- {
			if tags[i] == version1 && i > 0 {
				version2 = tags[i-1]
				break
			}
		}
		if version2 == "" {
			return "", fmt.Errorf("версия %s не найдена или предыдущая версия недоступна", version1)
		}
	}

	// Используем формат XML для вывода данных git log
	cmd := exec.Command("git", "log", "--pretty=format:<commit><hash>%H</hash><author>%an</author><email>%ae</email><date>%ad</date><message>%s</message></commit>", fmt.Sprintf("%s..%s", version2, version1))
	cmd.Dir = g.repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Ошибка при получении списка изменений между версиями %s и %s: %s\n%s", version1, version2, err, string(output))
		return "", err
	}

	commitsXML := "<commits>" + string(output) + "</commits>"

	// Преобразование XML в структуру Go
	var commits Commits
	if err := xml.Unmarshal([]byte(commitsXML), &commits); err != nil {
		return "", fmt.Errorf("ошибка при разборе XML: %v", err)
	}

	if format == JsonOutput {
		jsonOutput, err := json.MarshalIndent(commits, "", "  ")
		if err != nil {
			return "", err
		}
		return string(jsonOutput), nil
	} else {
		var result strings.Builder
		for _, commit := range commits.Commits {
			result.WriteString(fmt.Sprintf("%s\n", commit.Message))
		}
		return result.String(), nil
	}
}

// Функция для сравнения двух тегов
func compareTags(tag1, tag2 string) int {
	parts1 := strings.Split(strings.TrimPrefix(tag1, "v"), ".")
	parts2 := strings.Split(strings.TrimPrefix(tag2, "v"), ".")

	for i := 0; i < len(parts1); i++ {
		num1, _ := strconv.Atoi(parts1[i])
		num2, _ := strconv.Atoi(parts2[i])
		if num1 < num2 {
			return -1
		} else if num1 > num2 {
			return 1
		}
	}
	return 0
}

// Функция для инкрементации тега
func incrementTag(tag string, semVersionPart int) string {
	parts := strings.Split(strings.TrimPrefix(tag, "v"), ".")
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])
	switch semVersionPart {
	case SemVersionMajor:
		major++
	case SemVersionMinor:
		minor++
	case SemVersionPatch:
		patch++
	}
	return fmt.Sprintf("v%d.%d.%d", major, minor, patch)
}
