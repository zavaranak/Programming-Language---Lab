import http.client
import asyncio
import datetime
import re
import os
import uuid

total_bytes = 0
download_finished = False
filename = 'tempfile'

def findTitle():
    #for HTML file
    try:
        with open('tempfile', 'rb') as file:
            content = file.read()
        match = re.search(rb"<title>(.*?)</title>", content)
        title = match.group(1).strip() if match else None
        if (title==None): return None
        firstParsedTitle = title.decode('utf-8', errors='ignore') if title else None
        secondParsedTitle = handleTitleSyntax(firstParsedTitle)
        return secondParsedTitle
    except:
        return None

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
            print('MESSAGE: Invalid URL. Make sure you enter correct URL with http/https prefix')
        else:
            [httpConnnection,path] = parsedURL
            httpConnnection.request("GET", path) 
            response = httpConnnection.getresponse()
            contentType = response.headers.get("Content-Type")
            if response.status == 200:
                print("MESSAGE: Starting download...")
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
                
                extension = ''
                title =''
                if  "text/html" in contentType:
                    extension=".html"
                    original_name = findTitle()
                    title = original_name + extension if original_name else  str(uuid.uuid4())+extension
                elif "image" in contentType:
                    if "jpeg" in contentType or "jpg" in contentType:
                        extension = ".jpg"
                    elif "png" in contentType:
                        extension = ".png"
                    elif "gif" in contentType:
                        extension = ".gif"
                    elif "bmp" in contentType:
                        extension = ".bmp"
                    elif "webp" in contentType:
                        extension = ".webp"
                    elif "svg+xml" in contentType:
                        extension = ".svg"
                    else:
                        extension = ".img"  
                    title = str(uuid.uuid4()) + extension  
                elif "application" in contentType:
                    if "pdf" in contentType:
                        extension = ".pdf"
                    elif "msword" in contentType or "vnd.openxmlformats-officedocument.wordprocessingml.document" in contentType:
                            extension = ".docx"
                    elif "vnd.ms-excel" in contentType or "vnd.openxmlformats-officedocument.spreadsheetml.sheet" in contentType:
                            extension = ".xlsx"
                    elif "vnd.ms-powerpoint" in contentType or "vnd.openxmlformats-officedocument.presentationml.presentation" in contentType:
                        extension = ".pptx"
                    elif "zip" in contentType:
                        extension = ".zip"
                    else:
                        extension = ".doc" 
                    title = str(uuid.uuid4()) + extension 

                os.replace(filename, title)
                print(f"MESSAGE: File downloaded successfully as '{title}', size: {total_bytes} bytes")
            else:
                print(f"MESSAGE: Failed to download file. Status: {response.status} {response.reason}")
            httpConnnection.close()
            total_bytes = 0
    except:
        print(f"MESSAGE: Failed to download file. Please check your URL")
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
        command = input(">Enter the URL of file to download or Quit (q)\n>>> ")
        if command != 'q':
            await mainfunc(command)
            print(">RETURNING TO MENU:")
        elif command == "q":
            quit = True
    print(">Exiting the program.")

if __name__ == "__main__":
    asyncio.run(interface())