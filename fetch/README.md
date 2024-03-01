# fetch

given urls on stdin, will request them concurrently and save to the current directory, the file name for the url is actual file name on the server (e.g 'https://www.target.com/assets/js/main.js' will be saves as 'main.js')

inspired by @tomnomnom's fff :), its just simpler to use and creates file names differently, if you want more options using the tool you're much better off using fff.

TODO: fix the getFileName function
