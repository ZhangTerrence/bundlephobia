package main

import (
	"bundlephobia/types"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func findDependencies(package_ string, cache *map[string]*types.DependencyTree) *types.DependencyTree {
	dependencyTree := &types.DependencyTree{}

	response, err := http.Get(fmt.Sprintf("https://registry.npmjs.org/%s", package_))
	if err != nil {
		panic(err)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	var parsedBody types.Package
	err = json.Unmarshal(data, &parsedBody)
	if err != nil {
		panic(err)
	}

	packageExists := (*cache)[parsedBody.Name]
	if packageExists != nil {
		return (*cache)[parsedBody.Name]
	}

	dependencyTree.Name = parsedBody.Name
	dependencyTree.Dependencies = []*types.DependencyTree{}
	for name := range parsedBody.Versions[parsedBody.Tags.Latest].Dependencies {
		dependencyTree.Dependencies = append(dependencyTree.Dependencies, findDependencies(name, cache))
	}

	(*cache)[parsedBody.Name] = dependencyTree
	return dependencyTree
}

func printSet(dependencySet []string, source string) string {
	sb := strings.Builder{}

	sb.WriteString(source)
	sb.WriteString("\n")

	for i, name := range dependencySet {
		if i != len(dependencySet)-1 {
			sb.WriteString("\u251c\u2500\u2500 ")
			sb.WriteString(name)
			sb.WriteString("\n")
		} else {
			sb.WriteString("\u2514\u2500\u2500 ")
			sb.WriteString(name)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func printTree(dependencyTree *types.DependencyTree, sideLines []bool) string {
	if dependencyTree == nil {
		return ""
	}

	sb := strings.Builder{}

	sb.WriteString(dependencyTree.Name)
	sb.WriteString("\n")

	for i, dependency := range dependencyTree.Dependencies {
		for _, b := range sideLines {
			if b {
				sb.WriteString("    ")
			} else {
				sb.WriteString("\u2502   ")
			}
		}

		if i != len(dependencyTree.Dependencies)-1 {
			sb.WriteString("\u251c\u2500\u2500 ")
			sb.WriteString(printTree(dependency, append(sideLines, false)))
		} else {
			sb.WriteString("\u2514\u2500\u2500 ")
			sb.WriteString(printTree(dependency, append(sideLines, true)))
		}

	}

	return sb.String()
}

func getTreeSize(dependencyTree *types.DependencyTree) int {
	if dependencyTree == nil {
		return 0
	}

	size := 1

	for _, dependency := range dependencyTree.Dependencies {
		size += getTreeSize(dependency)
	}

	return size
}

func messageListener(session *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.ID == session.State.User.ID {
		return
	}

	if message.Content[:4] == "$set" {
		cache := &map[string]*types.DependencyTree{}
		package_ := message.Content[5:]
		findDependencies(package_, cache)
		dependencySet := make([]string, 0, len(*cache))
		for name := range *cache {
			if name != package_ {
				dependencySet = append(dependencySet, name)
			}
		}

		setString := printSet(dependencySet, package_)
		if len(setString) > 1993 {
			content := &discordgo.MessageSend{
				Content: "Tree too long (lol), here's a text file:",
				File:    &discordgo.File{Name: message.Content[5:] + ".txt", Reader: strings.NewReader(setString)},
			}
			_, err := session.ChannelMessageSendComplex(message.ChannelID, content)
			if err != nil {
				panic(err)
			}
		} else {
			content := "```\n" + setString + "```"
			_, err := session.ChannelMessageSend(message.ChannelID, content)
			if err != nil {
				panic(err)
			}
		}
	}

	if message.Content[:5] == "$tree" {
		cache := &map[string]*types.DependencyTree{}
		package_ := message.Content[6:]
		dependencyTree := findDependencies(package_, cache)
		treeString := printTree(dependencyTree, []bool{})

		if len(treeString) > 1993 {
			content := &discordgo.MessageSend{
				Content: "Tree too long (lol), here's a text file:",
				File:    &discordgo.File{Name: message.Content[6:] + ".txt", Reader: strings.NewReader(treeString)},
			}
			_, err := session.ChannelMessageSendComplex(message.ChannelID, content)
			if err != nil {
				panic(err)
			}
		} else {
			content := "```\n" + treeString + "```"
			_, err := session.ChannelMessageSend(message.ChannelID, content)
			if err != nil {
				panic(err)
			}
		}
	}

	if message.Content[:6] == "$stats" {
		cache := &map[string]*types.DependencyTree{}
		package_ := message.Content[7:]
		dependencyTree := findDependencies(package_, cache)

		formatString := "There are a total of %d unique packages installed.\n" +
			"Including duplicates, the total number of packages is %d. \n"
		treeSize := getTreeSize(dependencyTree)
		setSize := len(*cache)

		content := fmt.Sprintf(formatString, setSize, treeSize)
		_, err := session.ChannelMessageSend(message.ChannelID, content)
		if err != nil {
			panic(err)
		}
	}

	if message.Content[:4] == "$all" {
		cache := &map[string]*types.DependencyTree{}
		package_ := message.Content[5:]
		dependencyTree := findDependencies(package_, cache)
		dependencySet := make([]string, 0, len(*cache))
		for name := range *cache {
			if name != package_ {
				dependencySet = append(dependencySet, name)
			}
		}

		formatString := "There are a total of %d unique packages installed.\n" +
			"Including duplicates, the total number of packages is %d. \n"
		treeSize := getTreeSize(dependencyTree)
		setSize := len(*cache)

		treeString := printTree(dependencyTree, []bool{})
		setString := printSet(dependencySet, package_)
		completeString := "Tree\n" + treeString + "\nSet\n" + setString

		content := &discordgo.MessageSend{
			Content: fmt.Sprintf(formatString, setSize, treeSize),
			File:    &discordgo.File{Name: message.Content[5:] + ".txt", Reader: strings.NewReader(completeString)},
		}
		_, err := session.ChannelMessageSendComplex(message.ChannelID, content)
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	session, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	if err != nil {
		panic(err)
	}

	session.AddHandler(messageListener)

	session.Identify.Intents = discordgo.IntentGuildMessages

	err = session.Open()
	if err != nil {
		panic(err)
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	err = session.Close()
	if err != nil {
		panic(err)
	}
}
