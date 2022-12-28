# SendX-Backend-Assignment

## What the code does?
- Creates a server endpoint which takes the URL of a webpage, after getting the URL it fetches the webpage and downloads it as a file in the `files` folder locally.
- The server also accepts a retry limit as a parameter. It retries maximum upto 10 times or retry limit, whichever is lower, before either successfully downloading the webpage or marking the page as a failure.
- If the webpage has already been requested in the last 24 hours then it's served from the local cache.
- It also creates a pool of 5 workers that do the work of downloading the requested webpage to limit the requests.

## Sample Request

`GET: http://localhost:8080/pagesource?url=https://github.com&retry_limit=3`

## Sample Response

``{
    "url": "https://github.com",
    "filename": "files/2.html",
    "timestamp": 1672246680
}``
