package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strings"

	"github.com/schollz/progressbar/v3"
)

func main() {
	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	gitProtonLink := fmt.Sprintf("https://github.com/GloriousEggroll/proton-ge-custom/releases")
	nativeSteamPath := fmt.Sprintf("/home/%s/.steam/root/", user.Username)
	flatpakSteamPaths := fmt.Sprintf("/home/%s/.var/app/com.valvesoftware.Steam/data/Steam/", user.Username)

	var dir string
	if _, err := os.Stat(nativeSteamPath); err != nil {
		if os.IsNotExist(err) {
			os.Chdir(flatpakSteamPaths)
			fmt.Printf("\nОбнаружен Steam Flatpak!\n")
			dir = fmt.Sprintf("%scompatibilitytools.d", flatpakSteamPaths)
		}
	} else {
		os.Chdir(nativeSteamPath)
		fmt.Printf("\nОбнаружен Steam Native!\n")
		dir = fmt.Sprintf("%scompatibilitytools.d", nativeSteamPath)
	}

	os.Mkdir("compatibilitytools.d", os.ModePerm)
	os.Chdir("compatibilitytools.d")

	archiveProton := parcingProtonName(gitProtonLink)
	checkProtonInstall(archiveProton, dir)
	downloadFile(archiveProton, gitProtonLink)
	Uncompress(archiveProton, dir)

	fmt.Printf("Завершено.\n\n")
}

func parcingProtonName(gitProtonLink string) string {
	fmt.Printf("Получаем информацию о версии ProtonGE на github...\n")
	resp, err := http.Get(gitProtonLink)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	re := regexp.MustCompile("GE-Proton[0-9]-[0-9]{2}.tar.gz")
	matched := re.MatchString(string(body))
	if matched == false {
		re = regexp.MustCompile("GE-Proton[0-9]{2}-[0-9]{2}.tar.gz")
		matched = re.MatchString(string(body))
	}

	return re.FindString(string(body))
}

func checkProtonInstall(archiveProton string, dir string) {
	protonVersion := strings.SplitN(archiveProton, ".", 2)[0]
	dirsectories, _ := os.ReadDir(dir)

	for _, dirs := range dirsectories {
		if strings.Contains(protonVersion, dirs.Name()) {
			fmt.Printf("\nДанная версия Proton уже установлена. (%s)\n\n", protonVersion)
			os.Exit(0)
		}
	}
}

func downloadFile(archiveProton string, gitProtonLink string) {
	protonVersion := strings.SplitN(archiveProton, ".", 2)[0]
	fmt.Printf("Загрузка %s... Подождите...\n", protonVersion)
	protonURL := fmt.Sprintf("%s/download/%s/%s", gitProtonLink, protonVersion, archiveProton)

	req, _ := http.NewRequest("GET", protonURL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()

	f, _ := os.OpenFile(archiveProton, os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"",
	)
	io.Copy(io.MultiWriter(f, bar), resp.Body)
}

func Uncompress(protonRelease string, dir string) {
	fmt.Printf("Распаковываем архив %s...\n", protonRelease)
	exec.Command("tar", "-xf", protonRelease, "-C", dir).Output()
	os.Remove(protonRelease)
}
