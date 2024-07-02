package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/stealth"
	"io/ioutil"
	"strings"
)

const MaxItems = 6000

// Listing is an object to store the details of a car listing
// and enable fast marshalling to JSON
type Listing struct {
	Name         string `json:"name"`
	Price        string `json:"price"`
	Mileage      string `json:"mileage"`
	Interior     string `json:"interior"`
	Engine       string `json:"engine"`
	Transmission string `json:"transmission"`
	MPG          string `json:"mpg"`
	DriveTrain   string `json:"drive_train"`
	Exterior     string `json:"exterior"`
	BedLength    string `json:"bed_length"`
	Link         string `json:"link"`
}

// PageIterator is an object to store the current offset and page size
// as we iterate through the pages of listings
type PageIterator struct {
	startOffset int
	pageSize    int
}

// main initializes the browser and begins the scraping process
func main() {
	// Initialize Browser and navigate to page
	browser := rod.New().MustConnect().NoDefaultDevice()
	defer browser.MustClose()
	// Use a stealth browser to avoid rate limiting
	page := stealth.MustPage(browser)
	page.MustWaitStable()
	var allListings []*Listing
	// Initialize a page iterator with 100 listings per page
	iterator := &PageIterator{
		startOffset: 0,
		pageSize:    100,
	}

	// MaxItems is a hardcoded constant based on the current number of listings manually determined
	for iterator.startOffset < MaxItems {

		// Scrape the page and append the listings to the allListings slice
		listings, err := scrapePage(iterator, page)
		if err != nil {
			fmt.Println("Error scraping zip: ", err)
			// If there is an error, skip to the next page
			// but still write the current listings to the file
			allListings = append(allListings, listings...)
			iterator.startOffset += iterator.pageSize
			continue
		}
		// If there is no error, append the listings to the allListings slice
		// and increment the startOffset
		allListings = append(allListings, listings...)
		iterator.startOffset += iterator.pageSize
		// Write the listings to a file after each page
		// This leads to more I/O but ensures that the data is saved
		err = writeListingsToFile(allListings)
		if err != nil {
			fmt.Println("Error writing to file: ", err)
			continue
		}
	}
}

// scrapeAllListings operates on a single page, it retrieves all the listing elements and
// kicks off an individual scrape for each listing
func scrapeAllListings(page *rod.Page) []*Listing {
	var currentListings []*Listing
	pageItems, err := page.Elements(".inventory-listing")
	if err != nil {
		fmt.Println("items not found")
	}
	fmt.Printf("begin scraping %d items\n", len(pageItems))
	for _, item := range pageItems {
		iteratedList, err := clickIntoListingDetail(page, item)
		if err != nil {
			fmt.Println("Error clicking into listing detail: ", err)
			continue
		}
		currentListings = append(currentListings, iteratedList)
	}
	return currentListings

}

func writeListingsToFile(allListings []*Listing) error {
	// Marshal the list to JSON
	jsonData, err := json.Marshal(allListings)
	if err != nil {
		return err
	}
	// Write the JSON data to a file
	err = ioutil.WriteFile("out.json", jsonData, 0644)
	if err != nil {
		return err
	}
	return nil
}

func scrapePage(pageIterator *PageIterator, page *rod.Page) ([]*Listing, error) {
	targetURL := fmt.Sprintf(`https://www.kbb.com/cars-for-sale/all/toyota/tacoma?firstRecord=%d&numRecords=%d&newSearch=true&searchRadius=500`, pageIterator.startOffset, pageIterator.pageSize)
	page.MustNavigate(targetURL)
	page.MustWaitStable()
	// Scroll to the bottom to ensure all items are loaded
	footer := page.MustElement("#globalFooter")
	footer.MustScrollIntoView()

	// Begin a search
	var allListings []*Listing
	allListings = append(allListings, scrapeAllListings(page)...)
	fmt.Printf("Finished scraping %d-%d, retreived %d records\n", pageIterator.startOffset, pageIterator.pageSize+pageIterator.startOffset, len(allListings))
	return allListings, nil
}

func clickIntoListingDetail(page *rod.Page, element *rod.Element) (*Listing, error) {
	// Find the H2 element and click into it

	link, err := element.Element("a")
	if err != nil {
		return nil, err
	}
	href := link.MustAttribute("href")
	// Open a new tab and navigate to the listing
	if href == nil {
		return nil, errors.New("link to detail not found")
	}
	// create a new page to navigate to the listing
	newTab := stealth.MustPage(page.Browser())
	linkAddress := fmt.Sprintf(`https://www.kbb.com/%s`, *href)
	newTab.MustNavigate(linkAddress)
	if err != nil {
		return nil, err

	}
	newTab.MustWaitStable()
	defer newTab.MustClose()
	// retrieve the listing details
	details, err := newTab.Elements(".list-condensed")
	if err != nil {
		return nil, err
	}
	// Retrieve the name and price
	nameDiv := newTab.MustElement("h1")
	priceSpan := newTab.MustElementR("span", "\\$")
	listing := &Listing{
		Name:  nameDiv.MustText(),
		Price: priceSpan.MustText(),
		Link:  linkAddress,
	}
	// Loop through the details and assign them to the listing
	for _, detail := range details {
		if strings.Contains(detail.MustText(), "miles") {
			listing.Mileage = detail.MustText()
		}
		if strings.Contains(detail.MustText(), "Interior") {
			listing.Interior = detail.MustText()
		}
		if strings.Contains(detail.MustText(), "Engine") {
			listing.Engine = detail.MustText()
		}
		if strings.Contains(detail.MustText(), "Transmission") {
			listing.Transmission = detail.MustText()
		}
		if strings.Contains(detail.MustText(), "Highway") {
			listing.MPG = detail.MustText()
		}
		if strings.Contains(detail.MustText(), "drive") {
			listing.DriveTrain = detail.MustText()
		}
		if strings.Contains(detail.MustText(), "Exterior") {
			listing.Exterior = detail.MustText()
		}
		if strings.Contains(detail.MustText(), "Bed Length") {
			listing.BedLength = detail.MustText()
		}

	}
	return listing, nil
}
