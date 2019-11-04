# Roborock Oucher

## What is it?
Some time ago, Michael Reeves made [a video that went viral](https://www.youtube.com/watch?v=mvz3LRK263E), with a Roomba that was modded to scream when it hurts an obstacle. Michael removed many components from the Roomba, and that made the robot really funny but totally useless.

However, the Roborock, better known as Xiaomi Mi Vacuum Cleaner, already has all the components it needs to get the same result without any hardware modification, and without loosing any native functionality. So, we made up some little scripting that can be used on a rooted Roborock cleaner.

## What models does it work on?
It has been tested on:
- Xiaomi Mi Vacuum Cleaner gen1

It should work on any Roborock/Xiaomi Mi Vacuum Cleaner: if you successfully use it on other models please let us know by adding an issue so we can add it to the list. Don't be too scared to try if you don't have a compatible model: the script just reads a log file and doesn't make any modification to the system, so the worst thing that can happen is just that it doesn't work. The script, not the robot ;)

In all of this README I will talk about "Roborock" to mention the robot. This is just for simplicity: the instructions apply to all the compatible models.

## How do I install this?
First of all, you need to have a rooted Roborock. Please refer to [this wiki page](https://github.com/dgiese/dustcloud/wiki/VacuumRobots-manual-update-root-Howto) or search on the Internet about how to root your device. It's quite easy, but we won't offer support for this, sorry. :)

Download the oucher.sh and oucher.conf files from this repository, or just clone the entire repo.  
Then, you can edit the oucher.sh file and change the phrases that will be pronounced: a random one will be chosen each time. Make sure to set the correct language or the phrases will be pronunced in a strange way.

Then:
- Copy oucher.sh to the Roborock, in /usr/local/bin
- Copy oucher.conf to the Roborock, in /etc/init
- Log into SSH to the device
- Install espeak and alsa-utils: `apt-get update && apt-get install espeak alsa-utils && apt-get clean`
- Start the service: `service oucher start` (or just reboot the device)

All of this can be executed from the shell, in the folder where you downloaded the files:
```
scp oucher.sh root@192.168.1.33:/usr/local/bin
scp oucher.conf root@192.168.1.33:/etc/init
ssh root@192.168.1.33 apt-get update
ssh root@192.168.1.33 apt-get install espeak alsa-utils
ssh root@192.168.1.33 apt-get clean
ssh root@192.168.1.33 service oucher start
```
Just replace `192.168.1.33` with your Roborock IP.

Done! Just start a clean and wait for the first bump ;)

## How does it work?
The Roborock service logs everything that happens while cleaning in a file: `/run/shm/NAV_normal.log`. This includes bumps into obstacles. The scripts just follows the log file and invokes `espeak` everytime a bump occurs. A one-second delay is added after each pronounced phrase to avoid continuous screams.

## Are you planning to improve it?
Of course! We made this simple script just to make sure it was possible to achieve the goal. Short-term plans are:
- Add support for custom sounds: espeak isn't really good at screaming, and to be honest it's quite annoying. Real screams will fit better.
- Provide some real screams out-of-the-box, recorded by us. This will be funny ;)
- Improve the setup procedure, maybe by providing a deb file or a PPA to add to the Roborock

Anyway, we're sure you can get a great amount of fun with what already exists ;)

## I tried this and now my robot doesn't work! Shame on you!
Sorry for your loss :)  
Seriously: we're pretty confident it's not an issue with our script, since it really doesn't touch anything on the system.  
Most probably, you had some trouble with the root procedure. It's really hard to brick a Roborock, so maybe you'll find a solution if you search carefully on the dedicated channels. As said above, we're not giving support about the root procedure.

## I followed the procedure but the robot doesn't ouch
In this case, we're really happy to help! Just open an issue about it with as many details as you can, and we'll sort it out.