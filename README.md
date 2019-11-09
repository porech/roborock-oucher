# Roborock Oucher
![Roborock Oucher logo](https://i.ibb.co/5K16Hxr/oucher.jpg)

## What is it?
Some time ago, Michael Reeves made [a video that went viral](https://www.youtube.com/watch?v=mvz3LRK263E), with a Roomba that was modded to scream when it hurts an obstacle. Michael removed many components from the Roomba, and that made the robot really funny but totally useless.

However, the Roborock, better known as Xiaomi Mi Vacuum Cleaner, already has all the components it needs to get the same result without any hardware modification, and without loosing any native functionality. So, we made up a Golang application that can be used on a rooted Roborock cleaner.

## What models does it work on?
It has been tested on:
- Xiaomi Mi Vacuum Cleaner gen1
- Xiaomi Mi Vacuum Cleaner gen2
- Roborock S5

It should work on any Roborock/Xiaomi Mi Vacuum Cleaner: if you successfully use it on other models please let us know by adding an issue so we can add it to the list. Don't be too scared to try if you don't have a compatible model: the software just reads a log file and doesn't make any modification to the system, so the worst thing that can happen is just that it doesn't work. The screams, not the robot ;)

In all of this README I will talk about "Roborock" to mention the robot. This is just for simplicity: the instructions apply to all the compatible models.

## How do I install this?
First of all, you need to have a rooted Roborock. Please refer to [this wiki page](https://github.com/dgiese/dustcloud/wiki/VacuumRobots-manual-update-root-Howto) or search on the Internet about how to root your device. It's quite easy, but we won't offer support for this, sorry. :)

Download the `oucher` and `oucher.conf` files from this repository, or just clone the entire repo.

Then:
- If you already had a previous version, stop the oucher service: `service oucher stop`
- Copy `oucher` to the Roborock, in `/usr/local/bin`
- Copy `oucher.conf` to the Roborock, in `/etc/init`
- Log into SSH to the device
- Install espeak, sox and alsa-utils: `apt-get update && apt-get install espeak sox alsa-utils && apt-get clean`
- Start the service: `service oucher start` (or just reboot the device)

All of this can be executed from the shell, in the folder where you downloaded the files:
```bash
ssh root@192.168.1.33 service oucher stop
scp oucher root@192.168.1.33:/usr/local/bin
scp oucher.conf root@192.168.1.33:/etc/init
ssh root@192.168.1.33 apt-get -y update
ssh root@192.168.1.33 apt-get -y install espeak sox alsa-utils
ssh root@192.168.1.33 apt-get -y clean
ssh root@192.168.1.33 service oucher start
```
Just replace `192.168.1.33` with your Roborock IP.

If you're installing for the first time, the first command will return an error. That's normal, don't worry about it.

Done! Just start a clean and wait for the first bump ;)

## Can I customize the phrases?
Sure! Just customize the `oucher.yml` file and copy it to the Roborock, in the `/mnt/data/oucher` folder (you'll need to create it). From a shell:
```bash
ssh root@192.168.1.33 mkdir /mnt/data/oucher
scp oucher.yml root@192.168.1.33:/mnt/data/oucher
```
Just replace `192.168.1.33` with your Roborock IP.

Remember to restart the service with `service oucher restart` each time you make changes to the configuration, because the file is read on startup only.

## Can I use real screams?
Yes! You can create the /mnt/data/oucher/sounds folder (`mkdir -p /mnt/data/oucher/sounds`) and put some WAV files in there (no MP3, just WAV).  
If you prefer to put the files in a different folder, you can customize the `soundsPath` parameter in the config file.

The phrase will be chosen randomly on every bump, from the textual or WAV ones. If you want to use WAV files only, set the phrases to an empty array in the config file:
```yaml
phrases: []
```

Remember to restart the service with `service oucher restart` each time you add or remove a WAV file, because the list is loaded on startup only.

We're grouping some funny sound packs [on this page](http://www.linuxzogno.org/oucher-sounds/): they're made by Oucher users with samples found on the Internet. If you own copyright for some of the files and you don't like them to be there, please open an issue and we'll remove them.

## It's quite annoying...
You can set a delay in the configuration file. This way, the software will make sure that, after a scream is played, another one won't be played in the next N seconds. Set, for example, `delay: 10` and it will feel much better!

## What happens on a firmware upgrade?
A firmware upgrade will remove Oucher along with its dependencies and the root access. However, the `/mnt/data/oucher` folder is not deleted, so your configuration and custom sounds (if you put them here) are safe. You can root the device again and install Oucher back following the setup procedure above. Everything will work as before.

However, if you spent hours looking for the perfect sounds and phrases, we **strongly** recommend you to backup the config and WAV files, so you won't have to worry if for some reason you need to perform a factory reset.

## How can I remove it?
If you just want to disable the software but be able to enable it back easily, you can just set `enabled: false` in the configuration. This way, the software does absolutely nothing: after loading the configuration, it just sleeps, without reading the log file or anything else.

If you want to totally remove the software, just delete the `/usr/local/bin/oucher`and `/etc/init/oucher.conf` files. If you have a custom configuration, or custom sounds, also remove the `/mnt/data/oucher` folder.  
You won't also need espeak, sox and alsa-utils anymore, so you can remove them with `apt-get remove espeak sox alsa-utils` followed by an `apt-get autoremove` to uninstall their dependencies.

From the shell:
```bash
ssh root@192.168.1.33 rm /usr/local/bin/oucher /etc/init/oucher.conf
ssh root@192.168.1.33 rm -r /mnt/data/oucher
ssh root@192.168.1.33 apt-get remove espeak sox alsa-utils
ssh root@192.168.1.33 apt-get autoremove
```
Just replace `192.168.1.33` with your Roborock IP.

## How does it work?
The Roborock service logs everything that happens while cleaning in a file: `/run/shm/PLAYER_fprintf.log`. This includes bumps into obstacles. The software just follows the log file and, everytime a bump occurs, invokes `espeak` piped with `aplay` for text-to-speech, or `aplay` alone for WAVs. A semaphore avoids overlapped screams if multiple bumps occurr in a rapid sequence.

## I used an old version that looked for the oucher.yml file in /etc. Do I need to move it?
You're not forced to move it: the configuration file is also looked up from the /etc folder, like in previous versions. Anyway, we strongly suggest to put it in /mnt/data/oucher, so you won't lose it in case of firmware upgrade (see above).

## Are you planning to improve it?
Of course! Short-term plans are:
- Provide some real screams out-of-the-box, recorded by us. This will be funny ;)
- Improve the setup procedure, maybe by providing a deb file or a PPA to add to the Roborock

Anyway, we're sure you can get a great amount of fun with what already exists ;)

## I tried this and now my robot doesn't work! Shame on you!
Sorry for your loss :)
Seriously: we're pretty confident it's not an issue with our software, since it really doesn't touch anything on the system.
Most probably, you had some trouble with the root procedure. It's really hard to brick a Roborock, so maybe you'll find a solution if you search carefully on the dedicated channels. As said above, we're not giving support about the root procedure.

## I followed the procedure but the robot doesn't ouch
In this case, we're really happy to help! Just open an issue about it with as many details as you can, and we'll sort it out.

## I love it, can I offer you a coffee?
Wow, thanks! You can drop some Bitcoin to 35J2dPDFHweeB87LiYcHbhVmtgBsNrP4eH
