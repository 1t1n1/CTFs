#!/bin/sh

cd /home/chall/
su - chall -c "cd save2pdf && npm start > /tmp/save2pdf.log 2>&1 &"
su - chall -c "cd cardjacker && npm start"