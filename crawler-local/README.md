# Oggole Local Crawler

A simple web crawler using **Cheerio** that runs on your computer and sends data to your Oggole database.

## Why Local?

From the course: *"We don't want to run our web crawler on the same virtual machine as our server: This will result in competing resources."*

This runs on **your machine**, not your VM.

## Setup

```bash
npm install
```

## Usage

```bash
npm start
```

That's it! It will crawl Wikipedia pages and send them to your Oggole database.

## Configuration

Configure using environment variables in `.env` file or export them:

- `CRAWLER_API_KEY` - Your API key (required, must match server .env)
- `OGGOLE_API_URL` - API endpoint (default: http://localhost:8080/api/batch-pages)
- `START_URL` - Which page to start from (optional, edit in crawler.js)
- `MAX_PAGES` - How many pages to crawl (optional, edit in crawler.js)

Example `.env` file:
```
CRAWLER_API_KEY=your_key_here
OGGOLE_API_URL=http://localhost:9090/api/batch-pages
```

Note: START_URL and MAX_PAGES can be edited directly in crawler.js for development, but using environment variables is recommended for production use.

## How It Works

1. Crawls Wikipedia starting from `START_URL`
2. Uses **Cheerio** (jQuery-style selectors) to parse HTML efficiently
3. Extracts title, URL, and content from pages
4. Finds links to more pages (breadth-first crawling)
5. Sends batched data to your Oggole API
6. API inserts into database

## Technology

Built with **Cheerio** - a fast, lightweight HTML parser designed for web scraping:
- ⚡ Optimized for parsing and extracting data from HTML
- 🎯 jQuery-like syntax for familiar, readable selectors
- 📦 Minimal dependencies - perfect for Node.js crawlers
- 💪 Industry-standard tool for web scraping projects
