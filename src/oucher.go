package main

import (
	"bytes"
	"fmt"
	"github.com/papertrail/go-tail/follower"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var IsOuching bool = false
var ouchingMutex *sync.Mutex

type PhraseType int

const (
	Text PhraseType = iota
	Wav
)

type phrase struct {
	Type PhraseType
	Text string
}

type configuration struct {
	Enabled    bool     `mapstructure:enabled`
	Language   string   `mapstructure:language`
	Volume     int      `mapstructure:volume`
	SoundsPath string   `mapstructure:soundsPath`
	LogPaths   []string `mapstructure:logPaths`
	LogLevel   string   `mapstructure:logLevel`
	Phrases    []string `mapstructure:phrases`
	Delay      int      `mapstructure:delay`
	UseSox     bool
}

func main() {
	// initialize global pseudo random generator
	rand.Seed(time.Now().Unix())

	// Initialize ouching mutex
	ouchingMutex = &sync.Mutex{}

	// Set default configuration
	viper.SetDefault("enabled", true)
	viper.SetDefault("soundsPath", "/mnt/data/oucher/sounds")
	viper.SetDefault("logPaths", []string{"/run/shm/PLAYER_fprintf.log", "/run/shm/NAV_normal.log", "/run/shm/NAV_TRAP_normal.log"})
	viper.SetDefault("language", "en")
	viper.SetDefault("volume", 100)
	viper.SetDefault("phrases", []string{"Ouch!", "Argh!", "Hey, it hurts!"})
	viper.SetDefault("logLevel", "info")

	// Load the configuration file
	viper.SetConfigName("oucher")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/mnt/data/oucher")
	viper.AddConfigPath("/etc")
	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		log.Warnf("Error reading config file, using defaults: %s", err)
	}

	config := configuration{}
	err := viper.Unmarshal(&config)
	if err != nil {
		log.Warnf("Error parsing config file, using defaults: %s", err)
	}

	// Set the log level
	setLogLevel(config.LogLevel)

	// If disabled, just sleep
	if !config.Enabled {
		log.Debug("Disabled, waiting forever")
		for {
			time.Sleep(time.Minute)
		}
		return
	}

	// Initialize the phrases array
	var phrases []phrase

	// For each phrase, add it to the array
	for _, txtPhrase := range config.Phrases {
		log.Debugf("Phrase: %s", txtPhrase)
		newPhrase := phrase{
			Text,
			txtPhrase,
		}
		phrases = append(phrases, newPhrase)
	}

	// Search for wav files in the sounds path and add them to the list
	if dirExists(config.SoundsPath) {
		files, err := ioutil.ReadDir(config.SoundsPath)
		if err != nil {
			log.Warn("Can't read sounds directory: ", err)
		}

		for _, f := range files {
			if strings.HasSuffix(strings.ToLower(f.Name()), ".wav") {
				path := filepath.Join(config.SoundsPath, f.Name())
				log.Debugf("Sound: %s", path)
				newPhrase := phrase{
					Wav,
					path,
				}
				phrases = append(phrases, newPhrase)
			}
		}
	}

	// Perform some checks on the configuration
	if config.Volume > 100 {
		log.Warn("Volume is more than 100, setting in to 100")
		config.Volume = 100
	}
	if config.Volume < 0 {
		log.Warn("Volume is less than 0, setting it to 0")
		config.Volume = 0
	}

	// If volume is less then 100, check if Sox exists
	config.UseSox = false
	if config.Volume < 100 {
		if cmdExists("sox") {
			config.UseSox = true
		} else {
			log.Warn("Volume is less than 100 but sox is not installed, WAV files will play at full volume!")
		}
	}

	// For each log path, start the watching routine
	for _, FilePath := range config.LogPaths {
		log.Debugf("Starting log watching from %s", FilePath)
		go watchLog(FilePath, phrases, &config)
	}

	// Sit and wait
	select {}
}

// Keeps a file watched and calls processLine on every line
func watchLog(FilePath string, phrases []phrase, config *configuration) {
	for {
		// If the file does not exist, wait for a second and restart the loop
		if !fileExists(FilePath) {
			time.Sleep(1 * time.Second)
			continue
		}

		// Initialize the follower
		t, err := initFollower(FilePath)

		// If there was an error, log it, wait for a second and restart the loop
		if err != nil {
			log.Error(err)
			time.Sleep(1 * time.Second)
			continue
		}

		// Process every line
		for line := range t.Lines() {
			processLine(line.String(), phrases, config)
		}

		// If the follower exited with an error, log it, wait a second and restart the loop
		err = t.Err()
		if err != nil {
			log.Error(err)
			time.Sleep(1 * time.Second)
			continue
		}
	}
}

// Sets the log level
func setLogLevel(level string) {
	switch strings.ToLower(level) {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}

// Check if the file exists and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Check if the file exists and is a directory
func dirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// Check if a command exists
func cmdExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// Initialize the follower with the required parameters
func initFollower(filename string) (*follower.Follower, error) {
	t, err := follower.New(filename, follower.Config{
		Whence: io.SeekEnd,
		Offset: 0,
		Reopen: true,
	})
	return t, err
}

// Process a single line
func processLine(line string, phrases []phrase, config *configuration) {
	log.Tracef("Received line: %s", line)
	// If there is no ":Bumper" or "bumper 00 001 001 3" in the line, do nothing
	if !strings.Contains(line, ":Bumper") && !strings.Contains(line, "bumper 00 001 001 3") && !strings.Contains(line, ": Bumper") {
		return
	}

	// If there is "Curr:(0, 0, 0)" "bumper 00 001 001 3 0 0 0" in the line, do nothing (it's a bumper restore info)
	if strings.Contains(line, "Curr:(0, 0, 0)") || strings.Contains(line, "bumper 00 001 001 3 0 0 0") {
		return
	}

	log.Debugf("Received valid line: %s", line)

	// If the voices set is empty, do nothing
	if len(phrases) == 0 {
		log.Warn("No phrases or sounds!")
		return
	}

	// Lock the ouching mutex
	ouchingMutex.Lock()

	// If there is already an ouch in progress, do nothing
	if IsOuching {
		log.Debug("Already ouching, doing nothing")
		ouchingMutex.Unlock()
		return
	}

	// Set the ouching semaphore as true
	IsOuching = true

	// Unlock the ouching mutex
	ouchingMutex.Unlock()

	// Ouch!
	go ouch(phrases, config)
}

func ouch(phrases []phrase, config *configuration) {
	// Choose a random phrase
	sayPhrase := phrases[rand.Intn(len(phrases))]
	log.Debugf("Chosen phrase: %s", sayPhrase.Text)
	// Say the phrase
	if sayPhrase.Type == Text {
		eSpeak(sayPhrase.Text, config.Language, config.Volume)
	} else {
		if config.UseSox {
			soxaPlay(sayPhrase.Text, config.Volume)
		} else {
			aPlay(sayPhrase.Text)
		}
	}

	// If there is a delay set, wait before resetting the semaphore
	if config.Delay > 0 {
		time.Sleep(time.Duration(config.Delay) * time.Second)
	}

	// Unset the ouching semaphore
	ouchingMutex.Lock()
	IsOuching = false
	ouchingMutex.Unlock()
}

// Invoke the espeak command, piping it with aplay
func eSpeak(phrase, language string, volume int) {
	espeakCmd := exec.Command("espeak", "--stdout", "-a", fmt.Sprintf("%d", volume*2), "-v", language, phrase)
	aplayCmd := exec.Command("aplay", "-")
	log.Debugf("espeak command: %s", espeakCmd.String())
	log.Debugf("aplay command: %s", aplayCmd.String())
	r, w := io.Pipe()
	espeakCmd.Stdout = w
	aplayCmd.Stdin = r

	var b2 bytes.Buffer
	aplayCmd.Stdout = &b2

	espeakCmd.Start()
	aplayCmd.Start()
	espeakCmd.Wait()
	w.Close()
	aplayCmd.Wait()
	io.Copy(os.Stdout, &b2)
}

// Invoke the sox command, piping it with aplay
func soxaPlay(file string, volume int) {
	soxCmd := exec.Command("sox", "-v", fmt.Sprintf("%.2f", float32(volume)/100), file, "-t", "wav", "-")
	aplayCmd := exec.Command("aplay", "-")
	log.Debugf("sox command: %s", soxCmd.String())
	log.Debugf("aplay command: %s", aplayCmd.String())
	r, w := io.Pipe()
	soxCmd.Stdout = w
	aplayCmd.Stdin = r

	var b2 bytes.Buffer
	aplayCmd.Stdout = &b2

	soxCmd.Start()
	aplayCmd.Start()
	soxCmd.Wait()
	w.Close()
	aplayCmd.Wait()
	io.Copy(os.Stdout, &b2)
}

// Invoke the aplay command
func aPlay(file string) {
	aplayCmd := exec.Command("aplay", file)
	log.Debugf("aplay command: %s", aplayCmd.String())
	var b bytes.Buffer
	aplayCmd.Stdout = &b

	aplayCmd.Start()
	aplayCmd.Wait()
	io.Copy(os.Stdout, &b)
}
