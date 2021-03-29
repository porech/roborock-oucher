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

Download the `oucher`, `oucher.conf` and `S12oucher` files from this repository, or just clone the entire repo.

Then:
- If you already had a previous version, stop the oucher service: `service oucher stop`
- Copy `oucher` to the Roborock, in `/usr/local/bin`
- Copy the startupt script to the Roborock, in `/etc/init`. On the most recent models and firmware versions, the right startup script is `S12oucher`, while  on the other ones it's `oucher.conf`. Only one of them will work, the other one will be ignored, so if you're in doubt try one and then the other, or just copy both of them, it won't cause any damage.
- Log into SSH to the device
- If you're using a recent firmware version, edit the `/opt/rockrobo/rrlog/rrlog.conf` setting the LOG_LEVEL to 8 and reboot
- Make the oucher file executable by running: `chmod +x /usr/local/bin/oucher`
- Install espeak, sox and alsa-utils: `apt-get update && apt-get install espeak sox alsa-utils && apt-get clean`
- Reboot the device

All of this can be executed from the shell, in the folder where you downloaded the files:
```
export IP=192.168.1.33
ssh root@$IP service oucher stop
scp oucher root@$IP:/usr/local/bin
scp oucher.conf root@$IP:/etc/init
scp S12oucher root@$IP:/etc/init
ssh root@$IP chmod +x /usr/local/bin/oucher
ssh root@$IP sed -i -r 's/LOG_LEVEL=[0-9]*/LOG_LEVEL=8/' /opt/rockrobo/rrlog/rrlog.conf
ssh root@$IP apt-get -y update
ssh root@$IP apt-get -y install espeak sox alsa-utils
ssh root@$IP apt-get -y clean
ssh root@$IP reboot
```
Just replace `192.168.1.33` in the first command with your Roborock IP.

Depending on your model, some of the commands may return errors. Don't worry and go on with the next one.

Done! Just start a clean and wait for the first bump ;)

## The instructions above are not clear for me.
There is [a very well written tutorial](https://arner.github.io/posts/#install-the-oucher) by [Arner](https://arner.github.io/) you can follow. However I think that if you succeeded in rooting the device, that is the hard part, installing Oucher will be child's play!

## I'm on a newer/custom firmware without apt, how can I install the dependencies?
The awesome guys at the [dustcloud project](https://github.com/dgiese/dustcloud) put together all the needed binary dependencies in a single tgz file, which you can find in this repo as `oucher_deps.tgz`.
You should be fine by just copying it to the robot and uncompress it in the root folder:
```bash
tar xfv oucher_deps.tgz -C /
```

You can do this from your shell, in the folder where you downloaded the `oucher_deps.tgz` file:
```
export IP=192.168.1.33
scp oucher_deps.tgz root@$IP:/root
ssh root@$IP tar xfv /root/oucher_deps.tgz -C /
ssh root@$IP rm /root/oucher_deps.tgz
```
Just replace `192.168.1.33` in the first command with your Roborock IP.

## How can I build the executable by myself?
Just clone the repo, go into the `src` directory and run `./build.sh`. It will create the `oucher` file in the base project directory.

## It isn't working anymore on the latest firmware.
Please ensure you're using the latest binary available on the repository, and that you've set LOG_LEVEL to 8 in the `/opt/rockrobo/rrlog/rrlog.conf` file.

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
The Roborock service logs everything that happens while cleaning in a file: `/run/shm/PLAYER_fprintf.log` (or other files, depending on the model and firmware version). This includes bumps into obstacles. The software just follows the log file and, everytime a bump occurs, invokes `espeak` piped with `aplay` for text-to-speech, or `aplay` alone for WAVs. A semaphore avoids overlapped screams if multiple bumps occurr in a rapid sequence.

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
