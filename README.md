Sheep organizes a list of urls given through stdin into another with the query string parameters of the url page. This helps organizing and preparing your usage of SQLi tools.

Usage example:
cat list_of_urls | waybackurls | sheep | urldedup -s > sqli.txt
