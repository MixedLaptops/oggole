# Oggole Local Crawler

A simple web crawler that runs on your computer and sends data to your Oggole database.

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

Edit the top of `crawler.js` to change:

- `START_URL` - Which page to start from
- `MAX_PAGES` - How many pages to crawl
- `API_KEY` - Must match your `.env` file

## How It Works

1. Crawls Wikipedia starting from `START_URL`
2. Extracts title, URL, and content
3. Finds links to more pages
4. Sends everything to your Oggole API
5. API inserts into database

Simple!
