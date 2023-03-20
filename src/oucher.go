package main

import (
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
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

var IsOuching = false
var ouchingMutex *sync.Mutex

type PhraseType int

const (
	Wav PhraseType = iota
)

type phrase struct {
	Type PhraseType
	Text string
}

type configuration struct {
	Enabled    bool     `mapstructure:"enabled"`
	Volume     int      `mapstructure:"volume"`
	SoundsPath string   `mapstructure:"soundsPath"`
	LogPaths   []string `mapstructure:"logPaths"`
	LogLevel   string   `mapstructure:"logLevel"`
	Delay      int      `mapstructure:"delay"`

	OuchOnStart bool `mapstructure:"ouchOnStart"`
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
	viper.SetDefault("volume", 100)
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

	// Initialize the phdrases array
	var phrases []phrase

	// Search for wav files in the sounds path and add them to the list
	if !dirExists(config.SoundsPath) {
		log.Fatalf("sounds path %s does not exist", config.SoundsPath)
	}

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

	// Perform some checks on the configuration
	if config.Volume > 100 {
		log.Warn("Volume is more than 100, setting in to 100")
		config.Volume = 100
	}
	if config.Volume < 0 {
		log.Warn("Volume is less than 0, setting it to 0")
		config.Volume = 0
	}

	// If we should ouch on start, do it now
	if config.OuchOnStart {
		log.Debugf("Ouch on start is enabled, ouching now")
		go ouch(phrases, &config)
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

// allowed: strings that must be in the line for the event to be considered
var allowed = []string{
	":bumper",
	"bumper 00 001 001 3",
	": bumper",
	// S5 Max (https://github.com/porech/roborock-oucher/issues/14)
	"handletrap traphardwalkdetector  bumper counter",
}

// denied: strings that must NOT be in the line for the event to be considered
var denied = []string{
	"curr:(0, 0, 0)",
	"bumper 00 001 001 3 0 0 0",
	// https://github.com/porech/roborock-oucher/issues/17
	"subscribe",
	"suscribe",
	// https://github.com/porech/roborock-oucher/issues/17#issuecomment-991902784
	"bumperdatarecorder::bumperdatarecorderconfig",
}

func stringInArray(str string, arr []string) bool {
	for _, s := range arr {
		if strings.Contains(str, s) {
			return true
		}
	}
	return false
}

func isLineValid(line string) bool {
	lowercaseLine := strings.ToLower(line)
	if !stringInArray(lowercaseLine, allowed) {
		return false
	}
	if stringInArray(lowercaseLine, denied) {
		return false
	}
	return true
}

// Process a single line
func processLine(line string, phrases []phrase, config *configuration) {
	log.Tracef("Received line: %s", line)

	if !isLineValid(line) {
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

	// Play the file
	err := playSound(sayPhrase.Text, sayPhrase.Type, config.Volume)
	if err != nil {
		log.Errorf("cannot play file %s: %v", sayPhrase.Text, err)
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

func playSound(path string, phraseType PhraseType, volume int) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	streamer, format, err := wav.Decode(f)
	if err != nil {
		return fmt.Errorf("cannot decode file: %w", err)
	}
	defer streamer.Close()
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		return fmt.Errorf("cannot initialize speaker: %w", err)
	}
	defer speaker.Close()
	done := make(chan bool)
	log.Debugf("Volume: %d", volume)
	volumedStreamer := &effects.Volume{
		Streamer: streamer,
		Base:     2,
		Volume:   -5 + (float64(volume) / 20),
		Silent:   volume == 0,
	}
	speaker.Play(beep.Seq(volumedStreamer, beep.Callback(func() {
		done <- true
	})))
	<-done
	log.Debugf("Finished playing")
	return nil
}
