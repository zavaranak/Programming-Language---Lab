import http.client
import asyncio
import datetime
import re
import os

total_bytes = 0
download_finished = False
filename = 'tempfile'

def findTitle():
    with open('tempfile', 'rb') as file:
        content = file.read()
    match = re.search(rb"<title>(.*?)</title>", content)
    title = match.group(1).strip() if match else None
    if (title==None): return None
    firstParsedTitle = title.decode('utf-8', errors='ignore') if title else None
    secondParsedTitle = handleTitleSyntax(firstParsedTitle)
    return secondParsedTitle

def handleTitleSyntax(title):
    invalid_chars = r'[<>:"/\\|?*]'
    parsedTitle = re.sub(invalid_chars, '', title)
    return parsedTitle.strip()

def parseURL(url):
    try:
        host_and_path = url.split('://')[1].split("/")
        host = host_and_path[0]
        host_and_path.pop(0)
        path = "/" + ("/").join(host_and_path) if len(host_and_path)>1 else "/"
        if url.startswith("http://"):
            connection = http.client.HTTPConnection(host)
        elif url.startswith("https://"):
            connection = http.client.HTTPSConnection(host)
        else: 
            connection = None
        return [connection,path]
    except:
        return None

async def print_progress():
    global total_bytes, download_finished
    while not download_finished:
        if total_bytes>0:
            print(f">IN PROCESS: Downloaded {total_bytes} bytes at {datetime.datetime.now()}")
        await asyncio.sleep(1)
    return

async def download_file(url):
    try:
        global total_bytes, download_finished
        parsedURL = parseURL(url) 
        if (parsedURL==None):
            print('>>>MESSAGE: Invalid URL. Make sure you enter correct URL with http/https prefix')
        else:
            [httpConnnection,path] = parsedURL
            httpConnnection.request("GET", path) 
            response = httpConnnection.getresponse()
            if response.status == 200 and "text/html" in response.headers.get("Content-Type", ""):
                print(">>>MESSAGE: Starting download...")
                with open(filename, 'wb') as file:
                    original_name=None
                    buffer = bytearray(2048) #2MB per buffer reading
                    while True:
                        num_bytes_read = response.readinto(buffer)
                        total_bytes += num_bytes_read
                        if num_bytes_read == 0:
                            print('MESSAGE: Finished download')
                            download_finished = True 
                            break  
                        file.write(buffer[:num_bytes_read])            
                        await asyncio.sleep(0)

                original_name = findTitle()+'.html'
                os.replace(filename, original_name)
                print(f">>>MESSAGE: File downloaded successfully as {original_name} with total {total_bytes} bytes")
            else:
                print(f">>>MESSAGE: Failed to download file. Status: {response.status} {response.reason}")
            httpConnnection.close()
            total_bytes = 0
    except:
        print(f">>>MESSAGE: Failed to download file. Please check your URL")
    finally:
        download_finished = True
        return

async def mainfunc(url):
    global total_bytes,download_finished
    total_bytes = 0 
    download_finished=False
    await asyncio.gather(print_progress(),download_file(url))
    return

async def interface():
    print('>MENU:\n')
    quit = False
    while not quit:
        command = input(">Enter the URL of website to download or Quit (q)\n>>> ")
        if command != 'q':
            await mainfunc(command)
            print(">RETURNING TO MENU:")
        elif command == "q":
            quit = True
    print("Exiting the program.")

if __name__ == "__main__":
    asyncio.run(interface())