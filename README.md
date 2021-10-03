# Roborock Oucher
![Roborock Oucher logo](https://i.ibb.co/5K16Hxr/oucher.jpg)

## What is it?
Some time ago, Michael Reeves made [a video that went viral](https://www.youtube.com/watch?v=mvz3LRK263E), with a Roomba that was modded to scream when it bumps into something. Michael removed many components from the Roomba, and that made the robot really funny but totally useless.

However, the Roborock, better known as Xiaomi Mi Vacuum Cleaner, already has all the components it needs to get the same result without any hardware modification, and without loosing any native functionality. So, we made up a Golang application that can be used on a rooted Roborock cleaner.

## What models does it work on?
It has been tested on:
- Xiaomi Mi Vacuum Cleaner gen1
- Xiaomi Mi Vacuum Cleaner gen2
- Roborock S5
- Roborock S6

It should work on any Roborock/Xiaomi Mi Vacuum Cleaner: if you successfully use it on other models please let us know by adding an issue so we can add it to the list. Don't be too scared to try if you don't have a compatible model: the software just reads a log file and doesn't make any modification to the system, so the worst thing that can happen is that it doesn't work. The screams, not the robot ;)

In all of this README I will talk about "Roborock" to mention the robot. This is just for simplicity: the instructions apply to all the compatible models.

## How do I install this?
First of all, you need to have a rooted Roborock. Please refer to [this wiki page](https://github.com/dgiese/dustcloud/wiki/VacuumRobots-manual-update-root-Howto) or search on the Internet about how to root your device. It's quite easy, but we won't offer support for this, sorry. :)

Clone or download this GIT repository, and enter its folder from your terminal.

Then:
- If you already had a previous version, log into SSH to the device and stop the oucher service: `service oucher stop`
- Copy `oucher` to the Roborock, in `/usr/local/bin`
- Copy the startupt script to the Roborock, in `/etc/init`. On the most recent models and firmware versions, the right startup script is `S12oucher`, while  on the other ones it's `oucher.conf`. Only one of them will work, the other one will be ignored, so if you're in doubt try one and then the other, or just copy both of them, it won't cause any damage.
- Log into SSH to the device
- If you're using a recent firmware version, edit the `/opt/rockrobo/rrlog/rrlog.conf` setting the LOG_LEVEL to 8 and reboot
- Make the oucher file executable by running: `chmod +x /usr/local/bin/oucher`
- Create the `/mnt/data/oucher/sounds` directory
- Copy your WAV files inside it, or copy the three files in the `sounds` folder of the repository
- Reboot the device

All of this can be executed from the shell, from the folder in which you cloned the GIT repository:
```
export IP=192.168.1.33
ssh root@$IP service oucher stop
scp oucher root@$IP:/usr/local/bin
scp oucher.conf root@$IP:/etc/init
scp S12oucher root@$IP:/etc/init
ssh root@$IP chmod +x /usr/local/bin/oucher
ssh root@$IP sed -i -r 's/LOG_LEVEL=[0-9]*/LOG_LEVEL=8/' /opt/rockrobo/rrlog/rrlog.conf
ssh root@$IP mkdir -p /mnt/data/oucher/sounds
rsync -rav sounds/ root@$IP:/mnt/data/oucher/sounds/
ssh root@$IP reboot
```
Just replace `192.168.1.33` in the first command with your Roborock IP.

Depending on your model, some of the commands may return errors. Don't worry and go on with the next one.

Done! Just start a clean and wait for the first bump ;)

## How can I build the executable by myself?
You need to have [Golang already installed](https://golang.org/doc/install).

You also need the `arm-linux-gnueabihf-gcc` compiler, and `asound2` library for armhf. You can install them on Debian/Ubuntu by running:
```
sudo dpkg --add-architecture armhf
sudo apt update
sudo apt install -y gobjc-arm-linux-gnueabihf
sudo apt install -y libasound2-dev:armhf
```

Now clone the repo, go into the `src` directory and run `./build.sh`. It will create the `oucher` file in the base project directory.

## It isn't working anymore on the latest firmware.
Please ensure you're using the latest binary available on the repository, and that you've set LOG_LEVEL to 8 in the `/opt/rockrobo/rrlog/rrlog.conf` file.

## How can I customize the parameters?
Oucher does not need a configuration file to work, since all the parameters have a default.

However, if you need to, a `oucher.yml` file is present into the repository with the default configuration parameters.
You can edit it and then copy it to the Roborock, in the `/mnt/data/oucher` folder. From a shell:
```bash
scp oucher.yml root@192.168.1.33:/mnt/data/oucher
```
Just replace `192.168.1.33` with your Roborock IP.

Remember to restart the service with `service oucher restart` each time you make changes to the configuration, because the file is read on startup only.

## Can I use my own screams?
Yes! You can copy them into the /mnt/data/oucher/sounds folder (no MP3, just WAV).  
If you prefer to put the files in a different folder, you can customize the `soundsPath` parameter in the config file.

Remember to restart the service with `service oucher restart` each time you add or remove a WAV file, because the list is loaded on startup only.

We're grouping some funny sound packs [on this page](http://www.linuxzogno.org/oucher-sounds/): they're made by Oucher users with samples found on the Internet. If you own copyright for some of the files and you don't like them to be there, please open an issue and we'll remove them.

## It's quite annoying...
You can set a delay in the configuration file. This way, the software will make sure that, after a scream is played, another one won't be played in the next N seconds. Set, for example, `delay: 10` and it will feel much better!

## What happens on a firmware upgrade?
A firmware upgrade will remove Oucher along with its dependencies and the root access. However, the `/mnt/data/oucher` folder is not deleted, so your configuration and custom sounds (if you put them here) are safe. You can root the device again and install Oucher back following the setup procedure above. Everything will work as before.

However, if you spent hours looking for the perfect sounds, we **strongly** recommend you to backup the config and WAV files, so you won't have to worry if for some reason you need to perform a factory reset.

## How can I remove it?
If you just want to disable the software but be able to enable it back easily, you can just set `enabled: false` in the configuration. This way, the software does absolutely nothing: after loading the configuration, it just sleeps, without reading the log file or anything else.

If you want to totally remove the software, just delete the `/usr/local/bin/oucher`, `/etc/init/oucher.conf`, and `/etc/init/S12oucher` files. Also remove the `/mnt/data/oucher` folder.  

From the shell:
```
export IP=192.168.1.33
ssh root@$IP rm /usr/local/bin/oucher /etc/init/oucher.conf
ssh root@$IP rm -r /mnt/data/oucher
```
Just replace `192.168.1.33` in the first command with your Roborock IP.

## How does it work?
The Roborock service logs everything that happens while cleaning in a file: `/run/shm/PLAYER_fprintf.log` (or other files, depending on the model and firmware version). This includes bumps into obstacles. The software just follows the log file and, everytime a bump occurs, invokes `espeak` piped with `aplay` for text-to-speech, or `aplay` alone for WAVs. A semaphore avoids overlapped screams if multiple bumps occurr in a rapid sequence.

## I used an old version that looked for the oucher.yml file in /etc. Do I need to move it?
You're not forced to move it: the configuration file is also looked up from the /etc folder, like in previous versions. Anyway, we strongly suggest to put it in /mnt/data/oucher, so you won't lose it in case of firmware upgrade (see above).

## I used an old version that used a text-to-speech to play phrases, why am I forced to use WAV files now?
The old version used to require many dependencies (espeak, aplay, sox, ...) that were quite heavy considering the limited disk space, and they were not really trivial to install on some firmwares without APT.  
The current version now uses the `beep` Golang library, that uses `asound` under the hood. This is statically linked inside the executable, so that there's no need to install any dependency anymore.

Unfortunately, this isn't as easy with `espeak`, that was previously used for text-to-speech. Moreover, it wasn't that good, really.  
So I decided, at least for now, to drop support for it and to let Oucher just play WAV files.

However, you can still generate WAV files using your favorite text-to-speech software and use it.

Fox example, this is how files in the `sounds` folder of this repo were generated: 
```
espeak -w argh.wav "Argh!"
espeak -w ouch.wav "Ouch!"
espeak -w hey-it-hurts.wav "Hey, it hurts!"
```

You can set a language to `espeak` with the `-v` parameter, for example:
```
espeak -v it -w italian-voice.wav "This will be an Italian voice." 
```

If you really need the old version, there's still the `old-espeak-and-aplay` branch of this repository with it.

## I tried this and now my robot doesn't work! Shame on you!
Sorry for your loss :)
Seriously: we're pretty confident it's not an issue with our software, since it really doesn't touch anything on the system.
Most probably, you had some trouble with the root procedure. It's really hard to brick a Roborock, so maybe you'll find a solution if you search carefully on the dedicated channels. As said above, we're not giving support about the root procedure.

## I followed the procedure but the robot doesn't ouch / ouches at the wrong moment / woke me up at 3am by screaming for no reason.
In this case, we're really happy to help! Just open an issue about it with as many details as you can, and we'll sort it out. If you can, please copy the following files from the robot immediately after the unexpected behaviour occurrs, and attach them to the issue:
- /run/shm/PLAYER_fprintf.log
- /run/shm/NAV_normal.log
- /run/shm/NAV_TRAP_normal.log

## I love it, can I offer you a beer?
Wow, thanks! You can drop some Bitcoin to 35J2dPDFHweeB87LiYcHbhVmtgBsNrP4eH

## Credits
- The [dustcloud project](https://github.com/dgiese/dustcloud) for all the work on rooting the devices and documenting the procedure, nothing of this would be possible without their work.
- [ZVLDZ](https://github.com/zvldz/vacuum) for having added Oucher to his firmware builder, and for making those well-made startup scripts I included here.
- Everyone who provided feedback on his specific model/firmware version, Oucher has improved a lot thanks to this.

Did I forget to mention you? Sorry! Just open a issue :)
