# WebScrappingGo


# Command to install google-chrome

#install packages
sudo apt-get install -y curl unzip xvfb libxi6 libgconf-2-4 fonts-liberation
#get latest chrome
wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb

#install it
sudo apt install ./google-chrome-stable_current_amd64.deb

#test a screenshot
google-chrome --headless --disable-gpu --screenshot https://www.chromestatus.com/