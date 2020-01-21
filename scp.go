package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "scp",
		Short: "Access the SCP Archives",
		Long:  `Access the SCP Archives by specifying a number or URL.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			arg := args[0]
			isLink := strings.Index(arg, "http") == 0
			link := arg
			if !isLink {
				number, err := strconv.Atoi(arg)
				if err != nil {
					fmt.Print(err.Error())
					return
				}

				link = fmt.Sprintf("http://www.scp-wiki.net/scp-%03d", number)
			}
			scp, err := FetchArticle(link)
			if err != nil {
				fmt.Print(err.Error())
				return
			}
			fmt.Println(scp.Content)
		},
	}
)

type SCPArticle struct {
	Title   string
	Content string
	Rating  string
}

func FetchArticle(url string) (*SCPArticle, error) {
	res, err := http.Get(url)
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Got status %d", res.StatusCode)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}

	title := doc.Find("div#page-title").First().Text()
	contentBlock := doc.Find("div#page-content").First()
	contentBlock.Children().First().Remove()
	content := contentBlock.Text()
	content = strings.Trim(content, " \n")

	// get rating
	re := regexp.MustCompile(`(?i)object class: +(.*)`)
	matches := re.FindStringSubmatch(content)
	rating := "Unknown"
	if len(matches) > 1 {
		rating = strings.ToLower(matches[1])
		rating = strings.Title(rating)
	}

	// set rating colors
	type replaceFunc func(string, ...interface{}) string
	var replace replaceFunc
	switch rating {
	case "Safe":
		replace = color.GreenString
		break
	case "Euclid":
		replace = color.YellowString
		break
	case "Keter":
		replace = color.RedString
		break
	default:
		replace = color.BlackString
	}
	content = re.ReplaceAllString(content, fmt.Sprintf("Object class: %s", replace("%s", rating)))

	return &SCPArticle{
		title,
		content,
		rating,
	}, nil
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func main() {
	Execute()
}
