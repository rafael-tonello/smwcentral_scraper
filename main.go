package main

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/beeker1121/goque"

	"smwex/misc"
	"smwex/scraper"
)

const (
	_7zip               = "tools/7zip/7zz"
	_supermarioWorldRom = "assets/Super Mario World.smc"
	_flips              = "tools/flips/flips-linux"
)

var downloadqueue *goque.Queue
var processitemsqueue *goque.Queue
var pervar misc.PersistentVarStorage

// variable to determine if queue is current being consumed
var consumingdownloads bool = false
var consumingItemToProcess bool = false
var executableFolder string
var lastStatisticsTime int64 = time.Now().Unix()
var lastStatisticsLinksCount int64 = 0

func main() {
	//get the executable folder
	executableFolder, _ = os.Executable()
	executableFolder = path.Dir(executableFolder)

	os.Mkdir("./run", 0777)
	//create the directory ./queue
	//os.RemoveAll("./queue")
	os.Mkdir("./run/queue/downloads", 0777)
	os.Mkdir("./run/queue/toprocess", 0777)
	os.Mkdir("./run/downloads", 0777)
	os.Mkdir("./run/games", 0777)
	os.Mkdir("./run/tmp/patching", 0777)

	var err error
	downloadqueue, err = goque.OpenQueue("./run/queues/downloads")
	if err != nil {
		fmt.Println("Error creating queue: ", err)
		return
	}

	processitemsqueue, err = goque.OpenQueue("./run/queues/itemsToProcess")
	if err != nil {
		fmt.Println("Error creating queue: ", err)
		return
	}

	//add current downloaded files to the procesitemsqueue
	//files, _ := os.ReadDir("./run/downloads")
	//for _, file := range files {
	//	processitemsqueue.EnqueueString(file.Name())
	//}

	ConsumeProcessItemsQueue()
	ConsumeDownloadQueue()

	spr := scraper.NewScraper("run/smwex.db")
	spr.AddLimitingDomain("dl.smwcentral.net")
	spr.AddLimitingDomain("www.smwcentral.net")
	spr.AddLimitingDomain("smwcentral.net")
	spr.LinkFound = func(link string) scraper.LinkFoundResult {
		//cut string from the las '/'
		if strings.Contains(link, "/") {
			lastpart := link[strings.LastIndex(link, "/")+1:]
			if strings.Contains(lastpart, ".") {
				return scraper.LF_DO_NOT_PROCESS
			}
		}

		return scraper.LF_PROCEED
	}

	spr.OnPageDownload.Listen(func(data scraper.DownloadedPage) {
		//fmt.Println("Downloaded page: ", data.Url)
		if spr.VisitedLinks%100 == 0 {
			displayStatistics(spr, 100)
		}

		if strings.Contains(data.Content, "<td class=\"field\">Length:</td>") && strings.Contains(data.Content, "exit(s)") {
			//find the ".zip" in the file
			zipPos := strings.Index(data.Content, ".zip")
			if zipPos == -1 {
				return
			}
			//find the last "href" before the ".zip"
			hrefPos := strings.LastIndex(data.Content[:zipPos], "href")
			if hrefPos == -1 {
				return
			}

			//get the link
			link := data.Content[hrefPos+6 : zipPos+4]
			if strings.HasPrefix(link, "//") {
				link = "https:" + link
			}

			//add link to downloadqueue
			downloadqueue.EnqueueString(link)
			ConsumeDownloadQueue()

		}

	})

	go spr.Visit("https://www.smwcentral.net/?p=section&a=details&id=37734")
	go spr.Visit("https://www.smwcentral.net/?p=section&a=details&id=37734")
	go spr.Visit("https://www.smwcentral.net/?p=section&a=details&id=37734")
	go spr.Visit("https://www.smwcentral.net/?p=section&a=details&id=37734")
	go spr.Visit("https://www.smwcentral.net/?p=section&a=details&id=37734")

	for {
		time.Sleep(1 * time.Second)
	}

}

func ConsumeDownloadQueue() {
	ConsumeQueue(downloadqueue, &consumingdownloads, func(item string) {
		filename := DownloadZip(item)
		if filename != "" {
			_, _ = pervar.IncOrDecInt("stats.donloadedgames", 1)
			fmt.Println("Game zip downloaded: ", filename)

			//add file to processitemsqueue
			processitemsqueue.EnqueueString(filename)
			ConsumeProcessItemsQueue()
		}
	})
}

func ConsumeProcessItemsQueue() {
	ConsumeQueue(processitemsqueue, &consumingItemToProcess, func(filename string) {
		fmt.Println("Patching game file: ", filename)
		//force clear tmp folder
		os.RemoveAll("./run/tmp/patching")
		//os.Mkdir("./run/tmp/patching", 0777)
		fmt.Println("Creating tmp folder")
		_ = exec.Command("sh", "-c", "mkdir -p ./run/tmp/patching").Run()

		//unzip the file (using 7z)
		command := "\"./" + _7zip + "\" " +
			"e " +
			"\"./run/downloads/" + filename + "\" " +
			"-o\"./run/tmp/patching/\""

		fmt.Println("running command ", command)

		//run the command
		_ = exec.Command("sh", "-c", command).Run()

		//get all bps and ips files from /run/tmp/patching folder (do not user misc)
		files, _ := os.ReadDir("./run/tmp/patching")
		fmt.Println(len(files), " files found")
		for _, file := range files {
			fmt.Println("Processing file: ", file.Name())
			if strings.HasSuffix(file.Name(), ".bps") || strings.HasSuffix(file.Name(), ".ips") {
				//./flips-linux --apply [bpsfile] [supmeriorom] [output]
				destinationFile := "./run/games/" + strings.TrimSuffix(file.Name(), path.Ext(file.Name())) + ".smc"

				//skip if destination file already exists
				if _, err := os.Stat(destinationFile); err == nil {
					fmt.Println("Game already patched: ", file.Name())
					continue
				}

				command = "\"./" + _flips + "\" " +
					" --apply " +
					"\"./run/tmp/patching/" + file.Name() + "\" " +
					"\"./" + _supermarioWorldRom + "\" " +
					"\"" + destinationFile + "\""

				//run the command
				process := exec.Command("sh", "-c", command).Run()
				//get errorCode
				if process != nil {
					fmt.Println("Error applying patch: ", process)
				}

				fmt.Println("game patched: ", file.Name())
			}
		}
	})
}

func ConsumeQueue(queue *goque.Queue, controlvar *bool, callback func(string)) {
	if *controlvar {
		return
	}

	*controlvar = true
	go func() {
		for {
			item, err := queue.Dequeue()
			if err != nil {
				if err != goque.ErrEmpty {
					fmt.Println("Error dequeueing item: ", err)
				}
				*controlvar = false
				return
			}

			fmt.Println("Dequeued: ", item.ToString())
			callback(item.ToString())
		}
	}()
}

func DownloadZip(item string) string {
	fmt.Println("Downloading file: ", item)

	//get filename from url (last part of url)
	filename := item[strings.LastIndex(item, "/")+1:]
	//unescape the filename
	filename, _ = url.QueryUnescape(filename)

	err := misc.DownloadFile("./run/downloads/"+filename, item)
	if err != nil {
		fmt.Println("Error downloading file: ", err)
		return ""
	}

	fmt.Println("Downloaded file: ", filename)
	return filename
}

func displayStatistics(scrapper *scraper.Scraper, statisStepSize int) {

	fmt.Println("==============[ Statistics ]==============")

	fmt.Println("Valid links: ", scrapper.ValidLinks)
	fmt.Println("Visited links: ", scrapper.VisitedLinks)

	//get current time milisseconds
	currentTime := time.Now().Unix()

	//get the difference between the last time and the current time
	diff := currentTime - lastStatisticsTime
	lastStatisticsTime = currentTime
	if lastStatisticsLinksCount == 0 {
		lastStatisticsLinksCount = int64(scrapper.VisitedLinks)
		return
	}

	pagesPerSecond := (float64(scrapper.VisitedLinks) - float64(lastStatisticsLinksCount)) / float64(diff)

	lastStatisticsLinksCount = int64(scrapper.VisitedLinks)
	fmt.Println("Pages per second: ", pagesPerSecond)

}
