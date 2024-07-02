# Tacoma Data Scraper

The Tacoma data scraper is a small portfolio project that I've created to demonstrate a client-side scraper in golang.

I have done this scraper in `rod` which offers client side rendering options and UI interaction that can become necessary in more intensive scraping use cases.

That being said, this scraper is relatively simple, it established a connection to Kelley Blue Book to search for Used Toyota Tacomas and then visits each listing to 
collect details regarding the price and features of each vehicle. I am using these features in a predictive model that I use as a price estimator as I shop for
a used vehicle.
