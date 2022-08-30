package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/briandowns/spinner"
	"golang.org/x/term"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"time"
)

var (
	appVer      = "0.2"
	lstDot      = " • "
	shrcPath    = homeDir() + ".zshrc"
	profilePath = homeDir() + ".zprofile"
	arm64Path   = "/opt/homebrew/"
	amd64Path   = "/usr/local/"
	brewPrefix  = checkBrewPrefix()
	cmdAdmin    = "sudo"
	cmdPMS      = checkBrewPath()
	cmdGit      = "/usr/bin/git"
	pmsIns      = "install"
	//pmsReIn     = "reinstall"
	//pmsRm       = "remove"
	pmsAlt    = "--cask"
	pmsRepo   = "tap"
	cmdASDF   = checkASDFPath()
	p10kPath  = homeDir() + ".config/p10k/"
	p10kCache = homeDir() + ".cache/p10k-" + userName()
	tryLoop   = 0
	macLdBar  = spinner.New(spinner.CharSets[16], 50*time.Millisecond)
	runLdBar  = spinner.New(spinner.CharSets[11], 50*time.Millisecond)
	clrReset  = "\033[0m"
	clrRed    = "\033[31m"
	clrGreen  = "\033[32m"
	clrYellow = "\033[33m"
	clrBlue   = "\033[34m"
	clrPurple = "\033[35m"
	clrCyan   = "\033[36m"
	clrGrey   = "\033[37m"
)

func messageError(handling, msg, code string) {
	errOccurred := clrRed + "\nError occurred " + clrReset + "at "
	errMsgFormat := "\n" + clrRed + "Error >> " + clrReset + msg + " (" + code + ")\n"
	if handling == "fatal" || handling == "stop" {
		fmt.Print(errors.New(lstDot + "Fatal error" + errOccurred))
		log.Fatalln(errMsgFormat)
	} else if handling == "print" || handling == "continue" {
		log.Println(errMsgFormat)
	} else if handling == "panic" || handling == "detail" {
		fmt.Print(errors.New(lstDot + "Panic error" + errOccurred))
		panic(errMsgFormat)
	} else {
		fmt.Print(errors.New(lstDot + "Unknown error" + errOccurred))
		log.Fatalln(errMsgFormat)
	}
}

func checkError(err error, msg string) {
	if err != nil {
		messageError("fatal", msg, err.Error())
	}
}

func checkCmdError(err error, msg, pkg string) {
	if err != nil {
		messageError("print", msg+" "+clrYellow+pkg+clrReset, err.Error())
	}
}

func checkNetStatus() bool {
	getTimeout := 10000 * time.Millisecond
	client := http.Client{
		Timeout: getTimeout,
	}

	_, err := client.Get("https://9.9.9.9")
	if err != nil {
		return false
	}
	return true
}

func checkExists(path string) bool {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return true
	} else {
		return false
	}
}

func checkBrewPath() string {
	switch runtime.GOARCH {
	case "amd64":
		return amd64Path + "bin/brew"
	}
	return arm64Path + "bin/brew"
}

func checkBrewPrefix() string {
	switch runtime.GOARCH {
	case "amd64":
		return amd64Path
	}
	return arm64Path
}

func checkASDFPath() string {
	asdfPath := "opt/asdf/libexec/bin/asdf"
	switch runtime.GOARCH {
	case "amd64":
		return amd64Path + asdfPath
	}
	return arm64Path + asdfPath
}

func checkPassword() (string, bool) {
	for tryLoop < 3 {
		fmt.Print("Password:")
		bytePw, _ := term.ReadPassword(0)

		runLdBar.Suffix = " Checking password... "
		runLdBar.Start()

		tryLoop++
		strPw := string(bytePw)
		inputPw := exec.Command("echo", strPw)
		checkPw := exec.Command(cmdAdmin, "-Sv")
		checkPw.Env = os.Environ()
		checkPw.Stdout = os.Stdout

		checkPw.Stdin, _ = inputPw.StdoutPipe()

		_ = checkPw.Start()
		_ = inputPw.Run()
		errSudo := checkPw.Wait()
		if errSudo != nil {
			runLdBar.FinalMSG = clrRed + "Password check failed" + clrReset + "\n"
			runLdBar.Stop()
			if tryLoop < 3 {
				fmt.Println(errors.New(lstDot + "Sorry, try again."))
			} else if tryLoop >= 3 {
				fmt.Println(errors.New(lstDot + "3 incorrect password attempts."))
			}
		} else {
			runLdBar.Stop()
			if tryLoop == 1 {
				clearLine(tryLoop)
			} else {
				clearLine(tryLoop * 2)
			}
			return strPw, true
		}
	}
	return "", false
}

func checkPermission(runOpt, brewStatus string) bool {
	var expHeadMsg string
	var expMsg string

	expHeadMsg = clrYellow + "Need ROOT permission " + clrReset + "to " + clrYellow + "install " + clrPurple
	if runOpt == "1" || runOpt == "2" {
		if brewStatus == "Install" {
			fmt.Println(expHeadMsg + "Homebrew")
			return true
		} else {
			return false
		}
	} else {
		if brewStatus == "Install" {
			expMsg = expHeadMsg + "homebrew " + clrReset + "and " + clrPurple + "Applications" + clrReset + ": "
		} else if brewStatus == "Update" {
			expMsg = expHeadMsg + "Applications" + clrReset + ": "
		}

		if runOpt == "3" {
			fmt.Println(expMsg + "Loopback and BlackHole")
		} else if runOpt == "4" {
			fmt.Println(expMsg + "Java and BlackHole")
		} else if runOpt == "5" {
			fmt.Println(expMsg + "Java, BlackHole and VMware Fusion")
		} else if runOpt == "6" {
			fmt.Println(expMsg + "Java, Loopback and VMware Fusion")
		} else if runOpt == "7" {
			fmt.Println(expMsg + "Java, Loopback, VMware Fusion, Wireshark and Zenmap")
		}
		return true
	}
}

func needPermission(strPw string) {
	inputPw := exec.Command("echo", strPw)
	checkPw := exec.Command(cmdAdmin, "-Sv")
	checkPw.Env = os.Environ()
	checkPw.Stdout = os.Stdout

	checkPw.Stdin, _ = inputPw.StdoutPipe()
	_ = checkPw.Start()
	_ = inputPw.Run()
	errSudo := checkPw.Wait()
	checkError(errSudo, "Failed to run root permission")

	runRoot := exec.Command(cmdAdmin, "whoami")
	runRoot.Env = os.Environ()
	whoAmI, _ := runRoot.Output()

	if string(whoAmI) != "root\n" {
		msg := "Incorrect user, please check permission of sudo.\n" +
			lstDot + "It need sudo command of \"" + clrRed + "root" + clrReset + "\" user's permission.\n" +
			lstDot + "Working username: " + string(whoAmI)
		messageError("fatal", msg, "User")
	}
}

func homeDir() string {
	homeDirPath, err := os.UserHomeDir()
	checkError(err, "Failed to get home directory")
	return homeDirPath + "/"
}

func workingDir() string {
	workingDirPath, err := os.Getwd()
	checkError(err, "Failed to get working directory")
	return workingDirPath + "/"
}

func userName() string {
	workingUser, err := user.Current()
	checkError(err, "Failed to get current user")
	return workingUser.Username
}

func rebootOS(adminCode string) {
	runLdBar.Suffix = " Restarting macOS, please wait a moment ... "
	runLdBar.Start()
	time.Sleep(time.Second * 3)

	needPermission(adminCode)
	reboot := exec.Command(cmdAdmin, "shutdown", "-r", "now")
	if err := reboot.Run(); err != nil {
		runLdBar.FinalMSG = clrRed + "Error: " + clrReset
		runLdBar.Stop()
		fmt.Println(errors.New("failed to reboot macOS"))
	}

	runLdBar.FinalMSG = "⣾ Restarting macOS, please wait a moment ... "
	runLdBar.Stop()
}

func makeDir(dirPath string) {
	if checkExists(dirPath) != true {
		err := os.MkdirAll(dirPath, 0755)
		checkError(err, "Failed to make directory")
	}
}

func makeFile(filePath, fileContents string, fileMode int) {
	targetFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.FileMode(fileMode))
	checkError(err, "Failed to get file information to make new file from \""+filePath+"\"")

	defer func() {
		err := targetFile.Close()
		checkError(err, "Failed to finish make file to \""+filePath+"\"")
	}()

	_, err = targetFile.Write([]byte(fileContents))
	checkError(err, "Failed to fill in information to \""+filePath+"\"")
}

func removeFile(filePath string) {
	if checkExists(filePath) == true {
		err := os.Remove(filePath)
		checkError(err, "Failed to remove file \""+filePath+"\"")
	}
}

func linkFile(tgPath, destPath, linkType, permission, adminCode string) {
	if linkType == "hard" {
		if permission == "root" || permission == "sudo" || permission == "admin" {
			needPermission(adminCode)
			lnFile := exec.Command(cmdAdmin, "ln", "-sfn", tgPath, destPath)
			lnFile.Stderr = os.Stderr
			err := lnFile.Run()
			checkCmdError(err, "Add failed to hard link file", "\""+tgPath+"\"->\""+destPath+"\"")
		} else {
			if checkExists(tgPath) == true {
				if checkExists(destPath) == true {
					removeFile(destPath)
				}
				errHardlink := os.Link(tgPath, destPath)
				checkCmdError(errHardlink, "Add failed to hard link", "\""+tgPath+"\"->\""+destPath+"\"")
			}
		}
	} else if linkType == "symbolic" {
		if permission == "root" || permission == "sudo" || permission == "admin" {
			needPermission(adminCode)
			lnFile := exec.Command(cmdAdmin, "ln", "-sfn", tgPath, destPath)
			lnFile.Stderr = os.Stderr
			err := lnFile.Run()
			checkCmdError(err, "Add failed to symbolic link", "\""+tgPath+"\"->\""+destPath+"\"")
		} else {
			if checkExists(tgPath) == true {
				if checkExists(destPath) == true {
					removeFile(destPath)
				}
				errSymlink := os.Symlink(tgPath, destPath)
				checkCmdError(errSymlink, "Add failed to symbolic link\"", tgPath+"\"->\""+destPath+"\"")
				errLinkOwn := os.Lchown(destPath, os.Getuid(), os.Getgid())
				checkError(errLinkOwn, "Failed to change ownership of symlink \""+destPath+"\"")
			}
		}
	} else {
		messageError("fatal", "Invalid link type", "Link file")
	}
}

func downloadFile(filePath, urlPath string, fileMode int) {
	resp, err := http.Get(urlPath)
	checkError(err, "Failed to connect "+urlPath)

	defer func() {
		errBodyClose := resp.Body.Close()
		checkError(errBodyClose, "Failed to download from "+urlPath)
	}()

	rawFile, err := io.ReadAll(resp.Body)
	checkError(err, "Failed to read file information from "+urlPath)

	makeFile(filePath, string(rawFile), fileMode)
}

func appendContents(filePath, fileContents string, fileMode int) {
	targetFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, os.FileMode(fileMode))
	checkError(err, "Failed to get file information to append contents from \""+filePath+"\"")

	defer func() {
		err := targetFile.Close()
		checkError(err, "Failed to finish append contents to \""+filePath+"\"")
	}()

	_, err = targetFile.Write([]byte(fileContents))
	checkError(err, "Failed to append contents to \""+filePath+"\"")
}

func clearLine(line int) {
	for clear := 0; clear < line; clear++ {
		fmt.Printf("\033[1A\033[K")
	}
}

func brewUpdate() {
	updateHomebrew := exec.Command(cmdPMS, "update", "--auto-update")
	err := updateHomebrew.Run()
	checkCmdError(err, "Brew failed to", "update repositories")
}

func brewUpgrade() {
	brewUpdate()
	upgradeHomebrew := exec.Command(cmdPMS, "upgrade", "--greedy")
	err := upgradeHomebrew.Run()
	checkCmdError(err, "Brew failed to", "upgrade packages")
}

func brewRepository(repo string) {
	brewRepo := exec.Command(cmdPMS, pmsRepo, repo)
	err := brewRepo.Run()
	checkCmdError(err, "Brew failed to add ", repo)
}

func brewCleanup() {
	upgradeHomebrew := exec.Command(cmdPMS, "cleanup", "--prune=all", "-nsd")
	err := upgradeHomebrew.Run()
	checkCmdError(err, "Brew failed to", "cleanup old packages")
}

func brewRemoveCache() {
	upgradeHomebrew := exec.Command("rm", "-rf", "\"$(brew --cache)\"")
	err := upgradeHomebrew.Run()
	checkCmdError(err, "Brew failed to", "remove cache")
}

func brewInstall(pkg string) {
	if checkExists(brewPrefix+"Cellar/"+pkg) != true {
		brewUpdate()

		brewIns := exec.Command(cmdPMS, pmsIns, pkg)
		brewIns.Stderr = os.Stderr
		err := brewIns.Run()
		checkCmdError(err, "Brew failed to install", pkg)
	}
}

func brewCask(pkg, appName string) {
	if checkExists("/Applications/"+appName+".app") != true {
		brewUpdate()

		brewIns := exec.Command(cmdPMS, pmsIns, pmsAlt, pkg)
		err := brewIns.Run()
		checkCmdError(err, "Brew failed to install cask", pkg)
	}
}

func brewCaskSudo(pkg, appName, appPath, adminCode string) {
	if checkExists(appPath) != true {
		needPermission(adminCode)
		brewUpdate()

		brewIns := exec.Command(cmdPMS, pmsIns, pmsAlt, pkg)
		err := brewIns.Run()
		checkCmdError(err, "Brew failed to install cask", appName)
	}
}

func asdfInstall(plugin, version string) {
	if checkExists(homeDir()+".asdf/plugins/"+plugin) != true {
		asdfPlugin := exec.Command(cmdASDF, "plugin", "add", plugin)
		err := asdfPlugin.Run()
		checkCmdError(err, "ASDF-VM failed to add", plugin)
	}

	asdfIns := exec.Command(cmdASDF, pmsIns, plugin, version)
	errIns := asdfIns.Run()
	checkCmdError(errIns, "ASDF-VM", plugin)

	asdfGlobal := exec.Command(cmdASDF, "global", plugin, version)
	errConf := asdfGlobal.Run()
	checkCmdError(errConf, "ASDF-VM failed to install", plugin)
}

func addJavaHome(tgVer, destVer, adminCode string) {
	tgHead := brewPrefix + "opt/openjdk"
	tgTail := " /libexec/openjdk.jdk"
	lnDir := "/Library/Java/JavaVirtualMachines/openjdk"

	if checkExists(brewPrefix+"Cellar/openjdk"+tgVer) == true {
		linkFile(tgHead+tgVer+tgTail, lnDir+destVer+".jdk", "symbolic", "root", adminCode)
	}
}

func confA4s() {
	a4sPath := homeDir() + ".config/alias4sh"
	makeDir(a4sPath)
	makeFile(a4sPath+"/alias4.sh", "# ALIAS4SH", 0644)

	dlA4sPath := workingDir() + ".dev4mac-alias4sh.sh"

	downloadFile(dlA4sPath, "https://raw.githubusercontent.com/leelsey/Alias4sh/main/install.sh", 0644)

	installA4s := exec.Command("/bin/sh", dlA4sPath)
	if err := installA4s.Run(); err != nil {
		removeFile(dlA4sPath)
		checkError(err, "Failed to install Alias4sh")
	}

	removeFile(dlA4sPath)
}

func confG4s() {
	fmt.Println(clrCyan + "Git global configuration" + clrReset)

	setBranchMain := exec.Command(cmdGit, "config", "--global", "init.defaultBranch", "main")
	errBranchMain := setBranchMain.Run()
	checkError(errBranchMain, "Failed to change branch default name (master -> main)")
	fmt.Println(lstDot + "Main git branch default name changed master -> main")

	fmt.Println(lstDot + "Add user information to the global git config")
	consoleReader := bufio.NewScanner(os.Stdin)
	fmt.Print("  - User name: ")
	consoleReader.Scan()
	gitName := consoleReader.Text()
	fmt.Print("  - User email: ")
	consoleReader.Scan()
	gitEmail := consoleReader.Text()

	setUserName := exec.Command(cmdGit, "config", "--global", "user.name", gitName)
	errUserName := setUserName.Run()
	checkError(errUserName, "Failed to set git user name")
	setUserEmail := exec.Command(cmdGit, "config", "--global", "user.email", gitEmail)
	errUserEmail := setUserEmail.Run()
	checkError(errUserEmail, "Failed to set git user email")
	clearLine(3)
	fmt.Println(lstDot + "Saved git user name and email")

	ignoreDirPath := homeDir() + ".config/git/"
	ignorePath := ignoreDirPath + "gitignore_global"

	makeDir(ignoreDirPath)
	downloadFile(ignorePath, "https://raw.githubusercontent.com/leelsey/Git4set/main/gitignore-sample", 0644)

	setExcludesFile := exec.Command(cmdGit, "config", "--global", "core.excludesfile", ignorePath)
	errExcludesFile := setExcludesFile.Run()
	checkError(errExcludesFile, "Failed to set git global ignore file")

	fmt.Println(lstDot + "Complete setup \"gitignore_global\" in " + ignoreDirPath)
}

func installBrew() {
	insBrewPath := workingDir() + ".dev4mac-brew.sh"

	downloadFile(insBrewPath, "https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh", 0755)

	installHomebrew := exec.Command("/bin/bash", "-c", insBrewPath)
	installHomebrew.Env = append(os.Environ(), "NONINTERACTIVE=1")

	if err := installHomebrew.Run(); err != nil {
		removeFile(insBrewPath)
		checkError(err, "Failed to install Homebrew")
	}
	removeFile(insBrewPath)

	if checkExists(cmdPMS) == false {
		messageError("fatal", "Installed brew failed, please check your system", "Can't find Homebrew")
	}
}

func macBegin(adminCode string) {
	if checkExists(cmdPMS) == true {
		macLdBar.Suffix = " Updating homebrew... "
		macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "update homebrew!\n"
		macLdBar.Start()
	} else {
		needPermission(adminCode)

		macLdBar.Suffix = " Installing homebrew... "
		macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install and update homebrew!\n"
		macLdBar.Start()

		installBrew()
	}

	err := os.Chmod(brewPrefix+"share", 0755)
	checkError(err, "Failed to change permissions on "+brewPrefix+"share to 755")

	brewUpdate()
	brewRepository("homebrew/core")
	brewRepository("homebrew/cask")
	brewRepository("homebrew/cask-versions")
	brewUpgrade()

	macLdBar.Stop()
}

func macEnv() {
	macLdBar.Suffix = " Setting basic environment... "
	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "setup zsh environment!\n"
	macLdBar.Start()

	profileContents := "#    ___________  _____   ____  ______ _____ _      ______ \n" +
		"#   |___  /  __ \\|  __ \\ / __ \\|  ____|_   _| |    |  ____|\n" +
		"#      / /| |__) | |__) | |  | | |__    | | | |    | |__   \n" +
		"#     / / |  ___/|  _  /| |  | |  __|   | | | |    |  __|  \n" +
		"#    / /__| |    | | \\ \\| |__| | |     _| |_| |____| |____ \n" +
		"#   /_____|_|    |_|  \\_\\\\____/|_|    |_____|______|______|\n#\n" +
		"#  " + userName() + "’s zsh profile\n\n" +
		"# HOMEBREW\n" +
		"eval \"$(" + cmdPMS + " shellenv)\"\n\n"
	makeFile(profilePath, profileContents, 0644)

	shrcContents := "#   ______ _____ _    _ _____   _____\n" +
		"#  |___  // ____| |  | |  __ \\ / ____|\n" +
		"#     / /| (___ | |__| | |__) | |\n" +
		"#    / /  \\___ \\|  __  |  _  /| |\n" +
		"#   / /__ ____) | |  | | | \\ \\| |____\n" +
		"#  /_____|_____/|_|  |_|_|  \\_\\\\_____|\n#\n" +
		"#  " + userName() + "’s zsh run commands\n\n"
	makeFile(shrcPath, shrcContents, 0644)

	makeDir(homeDir() + ".config")
	makeDir(homeDir() + ".cache")

	macLdBar.Stop()
}

func macDependency(runOpt string) {
	macLdBar.Suffix = " Installing dependencies... "
	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install dependencies!\n"
	macLdBar.Start()

	brewInstall("pkg-config")
	brewInstall("ca-certificates")
	brewInstall("ncurses")
	brewInstall("openssl@3")
	brewInstall("openssl@1.1")
	brewInstall("readline")
	brewInstall("autoconf")
	brewInstall("automake")
	brewInstall("mpdecimal")
	brewInstall("utf8proc")
	brewInstall("m4")
	brewInstall("gmp")
	brewInstall("mpfr")
	brewInstall("gettext")
	brewInstall("jpeg-turbo")
	brewInstall("libtool")
	brewInstall("libevent")
	brewInstall("libffi")
	brewInstall("libtiff")
	brewInstall("libvmaf")
	brewInstall("libpng")
	brewInstall("libyaml")
	brewInstall("giflib")
	brewInstall("xz")
	brewInstall("gdbm")
	brewInstall("sqlite")
	brewInstall("lz4")
	brewInstall("zstd")
	brewInstall("hiredis")
	brewInstall("berkeley-db")
	brewInstall("asciidoctor")
	brewInstall("freetype")
	brewInstall("fontconfig")
	brewInstall("pcre")
	brewInstall("pcre2")

	shrcAppend := "# NCURSES\n" +
		"export PATH=\"" + brewPrefix + "opt/ncurses/bin:$PATH\"\n" +
		"export LDFLAGS=\"" + brewPrefix + "opt/ncurses/lib\"\n" +
		"export CPPFLAGS=\"" + brewPrefix + "opt/ncurses/include\"\n" +
		"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/ncurses/lib/pkgconfig\"\n\n" +
		"# OPENSSL-3\n" +
		"export PATH=\"" + brewPrefix + "opt/openssl@3/bin:$PATH\"\n" +
		"export LDFLAGS=\"-L" + brewPrefix + "opt/openssl@3/lib\"\n" +
		"export CPPFLAGS=\"-I" + brewPrefix + "opt/openssl@3/include\"\n" +
		"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/openssl@3/lib/pkgconfig\"\n\n" +
		"# OPENSSL-1.1\n" +
		"export PATH=\"" + brewPrefix + "opt/openssl@1.1/bin:$PATH\"\n" +
		"export LDFLAGS=\"-L" + brewPrefix + "opt/openssl@1.1/lib\"\n" +
		"export CPPFLAGS=\"-I" + brewPrefix + "opt/openssl@1.1/include\"\n" +
		"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/openssl@1.1/lib/pkgconfig\"\n\n"
	appendContents(shrcPath, shrcAppend, 0644)

	if runOpt != "2" && runOpt != "3" {
		brewInstall("ccache")
		brewInstall("gawk")
		brewInstall("tcl-tk")
		brewInstall("bash")
		brewInstall("zsh")
		brewInstall("perl")
		brewInstall("ruby")
		brewInstall("python@3.10")
		brewInstall("openjdk")
		brewInstall("ghc")
		brewInstall("cabal-install")
	}

	if runOpt == "6" || runOpt == "7" {
		brewInstall("krb5")
		brewInstall("libsodium")
		brewInstall("nettle")
		brewInstall("coreutils")
		brewInstall("ldns")
		brewInstall("isl")
		brewInstall("npth")
		brewInstall("gzip")
		brewInstall("bzip2")
		brewInstall("fop")
		brewInstall("little-cms2")
		brewInstall("imath")
		brewInstall("openldap")
		brewInstall("openexr")
		brewInstall("openjpeg")
		brewInstall("jpeg-xl")
		brewInstall("webp")
		brewInstall("rtmpdump")
		brewInstall("aom")
		brewInstall("screenresolution")
		brewInstall("gnu-getopt")
		brewInstall("brotli")
		brewInstall("bison")
		brewInstall("swig")
		brewInstall("re2c")
		brewInstall("icu4c")
		brewInstall("bdw-gc")
		brewInstall("guile")
		brewInstall("wxwidgets")
		brewInstall("sphinx-doc")
		brewInstall("docbook")
		brewInstall("docbook2x")
		brewInstall("docbook-xsl")
		brewInstall("xmlto")
		brewInstall("html-xml-utils")
		brewInstall("shared-mime-info")
		brewInstall("x265")
		brewInstall("oniguruma")
		brewInstall("zlib")
		brewInstall("glib")
		brewInstall("libgpg-error")
		brewInstall("libgcrypt")
		brewInstall("libunistring")
		brewInstall("libatomic_ops")
		brewInstall("libiconv")
		brewInstall("libmpc")
		brewInstall("libidn")
		brewInstall("libidn2")
		brewInstall("libssh2")
		brewInstall("libnghttp2")
		brewInstall("libxml2")
		brewInstall("libtasn1")
		brewInstall("libxslt")
		brewInstall("libavif")
		brewInstall("libzip")
		brewInstall("libde265")
		brewInstall("libheif")
		brewInstall("libksba")
		brewInstall("libusb")
		brewInstall("liblqr")
		brewInstall("libomp")
		brewInstall("libassuan")
		brewInstall("p11-kit")
		brewInstall("gnutls")
		brewInstall("gd")
		brewInstall("ghostscript")
		brewInstall("imagemagick")
		brewInstall("pinentry")
		brewInstall("gnupg")

		shrcAppend := "# KRB5\n" +
			"export PATH=\"" + brewPrefix + "opt/krb5/bin:$PATH\"\n" +
			"export PATH=\"" + brewPrefix + "opt/krb5/sbin:$PATH\"\n" +
			"export LDFLAGS=\"" + brewPrefix + "opt/krb5/lib\"\n" +
			"export CPPFLAGS=\"" + brewPrefix + "opt/krb5/include\"\n" +
			"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/krb5/lib/pkgconfig\"\n\n" +
			"# COREUTILS\n" +
			"export PATH=\"" + brewPrefix + "opt/coreutils/libexec/gnubin:$PATH\"\n\n" +
			"export PATH=\"" + brewPrefix + "opt/gnu-getopt/bin:$PATH\"\n\n" +
			"# BZIP2\n" +
			"export PATH=\"" + brewPrefix + "opt/bzip2/bin:$PATH\"\n" +
			"export LDFLAGS=\"" + brewPrefix + "opt/bzip2/lib\"\n" +
			"export CPPFLAGS=\"" + brewPrefix + "opt/bzip2/include\"\n\n" +
			"# GNU GETOPT\n" +
			"# BISON\n" +
			"export PATH=\"" + brewPrefix + "opt/bison/bin:$PATH\"\n" +
			"export LDFLAGS=\"" + brewPrefix + "opt/bison/lib\"\n\n" +
			"# ICU4C\n" +
			"export PATH=\"" + brewPrefix + "opt/icu4c/bin:$PATH\"\n" +
			"export PATH=\"" + brewPrefix + "opt/icu4c/sbin:$PATH\"\n" +
			"export LDFLAGS=\"" + brewPrefix + "opt/icu4c/lib\"\n" +
			"export CPPFLAGS=\"" + brewPrefix + "opt/icu4c/include\"\n" +
			"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/icu4c/lib/pkgconfig\"\n\n" +
			"# DOCBOOK" +
			"export XML_CATALOG_FILES=\"" + brewPrefix + "etc/xml/catalog\"\n\n" +
			"# ZLIB\n" +
			"export LDFLAGS=\"" + brewPrefix + "opt/zlib/lib\"\n" +
			"export CPPFLAGS=\"" + brewPrefix + "opt/zlib/include\"\n" +
			"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/zlib/lib/pkgconfig\"\n\n" +
			"# LIBICONV\n" +
			"export PATH=\"" + brewPrefix + "opt/libiconv/bin:$PATH\"\n" +
			"export LDFLAGS=\"" + brewPrefix + "opt/libiconv/lib\"\n" +
			"export CPPFLAGS=\"" + brewPrefix + "opt/libiconv/include\"\n\n" +
			"# LIBXML2\n" +
			"export PATH=\"" + brewPrefix + "opt/libxml2/bin:$PATH\"\n" +
			"export LDFLAGS=\"" + brewPrefix + "opt/libxml2/lib\"\n" +
			"export CPPFLAGS=\"" + brewPrefix + "opt/libxml2/include\"\n" +
			"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/libxml2/lib/pkgconfig\"\n\n" +
			"# LIBXSLT\n" +
			"export PATH=\"" + brewPrefix + "opt/libxslt/bin:$PATH\"\n" +
			"export LDFLAGS=\"" + brewPrefix + "opt/libxslt/lib\"\n" +
			"export CPPFLAGS=\"" + brewPrefix + "opt/libxslt/include\"\n" +
			"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/libxslt/lib/pkgconfig\"\n\n"
		appendContents(shrcPath, shrcAppend, 0644)
	}

	macLdBar.Stop()
}

func macLanguage(runOpt, adminCode string) {
	macLdBar.Suffix = " Installing computer programming language... "
	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install languages!\n"
	macLdBar.Start()

	shrcAppend := "# CCACHE\n" +
		"export PATH=\"" + brewPrefix + "opt/ccache/libexec:$PATH\"\n\n" +
		"# TCL-TK\n" +
		"export PATH=\"" + brewPrefix + "opt/tcl-tk/bin:$PATH\"\n" +
		"export LDFLAGS=\"" + brewPrefix + "opt/tcl-tk/lib\"\n" +
		"export CPPFLAGS=\"" + brewPrefix + "opt/tcl-tk/include\"\n" +
		"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/tcl-tk/lib/pkgconfig\"\n\n" +
		"# RUBY\n" +
		"export PATH=\"" + brewPrefix + "opt/ruby/bin:$PATH\"\n" +
		"export LDFLAGS=\"" + brewPrefix + "opt/ruby/lib\"\n" +
		"export CPPFLAGS=\"" + brewPrefix + "opt/ruby/include\"\n" +
		"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/ruby/lib/pkgconfig\"\n\n"
	appendContents(shrcPath, shrcAppend, 0644)

	if runOpt == "4" || runOpt == "5" || runOpt == "6" || runOpt == "7" {
		brewInstall("php")
		brewInstall("openjdk@8")
		brewInstall("openjdk@11")
		brewInstall("openjdk@17")
		addJavaHome("", "", adminCode)
		addJavaHome("@17", "-17", adminCode)
		addJavaHome("@11", "-11", adminCode)
		addJavaHome("@8", "-8", adminCode)
	}

	if runOpt == "3" || runOpt == "4" || runOpt == "5" {
		brewInstall("nvm")
		brewInstall("pyenv")
		brewInstall("pyenv-virtualenv")

		shrcAppend := "# NVM\n" +
			"export NVM_DIR=\"$HOME/.nvm\"\n" +
			"[ -s \"" + brewPrefix + "opt/nvm/nvm.sh\" ] && source \"" + brewPrefix + "opt/nvm/nvm.sh\"\n" +
			"[ -s \"" + brewPrefix + "opt/nvm/etc/bash_completion.d/nvm\" ] && source \"" +
			brewPrefix + "opt/nvm/etc/bash_completion.d/nvm\"\n\n" +
			"# PYENV" +
			"export PYENV_ROOT=\"$HOME/.pyenv\"\n" +
			"export PATH=\"$PYENV_ROOT/bin:$PATH\"\n" +
			"eval \"$(pyenv init --path)\"\n" +
			"eval \"$(pyenv init -)\"\n\n"
		appendContents(shrcPath, shrcAppend, 0644)

		nvmIns := exec.Command("nvm", pmsIns, "--lts")
		nvmIns.Stderr = os.Stderr
		err := nvmIns.Run()
		checkCmdError(err, "NVM failed to install", "LTS")
	} else if runOpt == "6" || runOpt == "7" {
		brewInstall("llvm")
		brewInstall("gcc") // fortran
		brewInstall("go")
		brewInstall("rust")
		brewInstall("node")
		brewInstall("lua")
		brewInstall("groovy")
		brewInstall("kotlin")
		brewInstall("scala")
		brewInstall("clojure")
		brewInstall("erlang")
		brewInstall("elixir")
		brewInstall("typescript")
		brewInstall("haskell-stack")
		brewInstall("haskell-language-server")
		brewInstall("stylish-haskell")
		stackIns := exec.Command("stack", pmsIns, "cabal-install")
		err := stackIns.Run()
		checkCmdError(err, "Stack(haskell) failed to install", "cabal")
	}

	macLdBar.Stop()
}

func macServer() {
	macLdBar.Suffix = " Installing developing tools for server... "
	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install servers!\n"
	macLdBar.Start()

	brewInstall("httpd")
	brewInstall("tomcat")
	brewInstall("nginx")

	macLdBar.Stop()
}

func macDatabase() {
	macLdBar.Suffix = " Installing developing tools for database... "
	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install databases!\n"
	macLdBar.Start()

	shrcAppend := "# SQLITE3\n" +
		"export PATH=\"" + brewPrefix + "opt/sqlite/bin:$PATH\"\n" +
		"export LDFLAGS=\"" + brewPrefix + "opt/sqlite/lib\"\n" +
		"export CPPFLAGS=\"" + brewPrefix + "opt/sqlite/include\"\n" +
		"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/sqlite/lib/pkgconfig\"\n\n"
	appendContents(shrcPath, shrcAppend, 0644)

	brewInstall("sqlite-analyzer")
	brewInstall("postgresql")
	brewInstall("mysql")
	brewInstall("redis")
	brewRepository("mongodb/brew")
	brewInstall("mongodb-community")

	macLdBar.Stop()
}

func macDevVM() {
	macLdBar.Suffix = " Installing developer tools version management tool with plugin... "
	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install ASDF-VM with languages!\n"
	macLdBar.Start()

	brewInstall("asdf")

	shrcAppend := "# ASDF VM\n" +
		"source " + brewPrefix + "opt/asdf/libexec/asdf.sh\n" +
		"#source " + homeDir() + ".asdf/plugins/java/set-java-home.zsh\n\n"
	appendContents(shrcPath, shrcAppend, 0644)

	asdfrcContents := "#              _____ _____  ______  __      ____  __ \n" +
		"#       /\\    / ____|  __ \\|  ____| \\ \\    / /  \\/  |\n" +
		"#      /  \\  | (___ | |  | | |__ ____\\ \\  / /| \\  / |\n" +
		"#     / /\\ \\  \\___ \\| |  | |  __|_____\\ \\/ / | |\\/| |\n" +
		"#    / ____ \\ ____) | |__| | |         \\  /  | |  | |\n" +
		"#   /_/    \\_\\_____/|_____/|_|          \\/   |_|  |_|\n#\n" +
		"#  " + userName() + "’s ASDF-VM run commands\n\n" +
		"# JAVA\n" +
		"java_macos_integration_enable = yes\n\n"
	makeFile(homeDir()+".asdfrc", asdfrcContents, 0644)

	asdfInstall("perl", "latest")
	//asdfInstall("ruby", "latest")   // error
	//asdfInstall("python", "latest") // error
	asdfInstall("java", "openjdk-11.0.2") // JDK LTS 11
	asdfInstall("java", "openjdk-17.0.2") // JDK LTS 17
	asdfInstall("rust", "latest")
	asdfInstall("golang", "latest")
	asdfInstall("nodejs", "latest")
	asdfInstall("lua", "latest")
	//asdfInstall("php", "latest") // error
	asdfInstall("groovy", "latest")
	asdfInstall("kotlin", "latest")
	asdfInstall("scala", "latest")
	asdfInstall("clojure", "latest")
	//asdfInstall("erlang", "latest") // error
	asdfInstall("elixir", "latest")
	//asdfInstall("haskell", "latest") // error
	asdfInstall("gleam", "latest")

	asdfReshim := exec.Command(cmdASDF, "reshim")
	err := asdfReshim.Run()
	checkCmdError(err, "ASDF failed to", "reshim")

	macLdBar.Stop()
}

func macTerminal(runOpt string) {
	macLdBar.Suffix = " Installing zsh with useful tools... "
	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install and configure for terminal!\n"
	macLdBar.Start()

	confA4s()
	brewInstall("zsh-completions")
	brewInstall("zsh-syntax-highlighting")
	brewInstall("zsh-autosuggestions")
	brewInstall("z")
	brewInstall("tree")
	brewRepository("romkatv/powerlevel10k")
	brewInstall("romkatv/powerlevel10k/powerlevel10k")

	makeFile(homeDir()+".z", "", 0644)
	makeDir(p10kPath)
	makeDir(p10kCache)

	if runOpt == "5" || runOpt == "6" || runOpt == "7" {
		brewInstall("fzf")
		brewInstall("tmux")
		brewInstall("tmuxinator")
		brewInstall("neofetch")

		dliTerm2Conf := homeDir() + "Library/Preferences/com.googlecode.iterm2.plist"
		downloadFile(dliTerm2Conf, "https://raw.githubusercontent.com/leelsey/ConfStore/main/iterm2/iTerm2.plist", 0644)
	}

	dlP10kTerm := p10kPath + "p10k-term.zsh"
	downloadFile(dlP10kTerm, "https://raw.githubusercontent.com/leelsey/ConfStore/main/p10k/p10k-devsimple.zsh", 0644)

	if runOpt == "2" || runOpt == "3" || runOpt == "4" {
		profileAppend := "# POWERLEVEL10K\n" +
			"source " + brewPrefix + "opt/powerlevel10k/powerlevel10k.zsh-theme\n" +
			"if [[ -r \"${XDG_CACHE_HOME:-" + p10kCache + "}/p10k-instant-prompt-${(%):-%n}.zsh\" ]]; then\n" +
			"  source \"${XDG_CACHE_HOME:-" + p10kCache + "}/p10k-instant-prompt-${(%):-%n}.zsh\"\n" +
			"fi\n" +
			"[[ ! -f " + p10kPath + "p10k-terminal.zsh ]] || source " + p10kPath + "p10k-terminal.zsh\n\n"
		appendContents(profilePath, profileAppend, 0644)
	} else if runOpt == "5" || runOpt == "6" || runOpt == "7" {
		dlP10kiTerm2 := p10kPath + "p10k-iterm2.zsh"
		dlP10kTMUX := p10kPath + "p10k-tmux.zsh"
		dlP10kEtc := p10kPath + "p10k-etc.zsh"

		downloadFile(dlP10kiTerm2, "https://raw.githubusercontent.com/leelsey/ConfStore/main/p10k/p10k-devwork.zsh", 0644)
		downloadFile(dlP10kTMUX, "https://raw.githubusercontent.com/leelsey/ConfStore/main/p10k/p10k-devhelp.zsh", 0644)
		downloadFile(dlP10kEtc, "https://raw.githubusercontent.com/leelsey/ConfStore/main/p10k/p10k-devbegin.zsh", 0644)

		profileAppend := "# ZSH\n" +
			"export SHELL=zsh\n\n" +
			"# POWERLEVEL10K\n" +
			"source " + brewPrefix + "opt/powerlevel10k/powerlevel10k.zsh-theme\n" +
			"if [[ -r \"${XDG_CACHE_HOME:-" + p10kCache + "}/p10k-instant-prompt-${(%):-%n}.zsh\" ]]; then\n" +
			"  source \"${XDG_CACHE_HOME:-" + p10kCache + "}/p10k-instant-prompt-${(%):-%n}.zsh\"\n" +
			"fi\n" +
			"if [[ -d /Applications/iTerm.app ]]; then\n" +
			"  if [[ $TERM_PROGRAM = \"Apple_Terminal\" ]]; then\n" +
			"    [[ ! -f " + p10kPath + "p10k-term.zsh ]] || source " + p10kPath + "p10k-term.zsh\n" +
			"  elif [[ $TERM_PROGRAM = \"iTerm.app\" ]]; then\n" +
			"    echo ''; neofetch --bold off\n" +
			"    [[ ! -f " + p10kPath + "p10k-iterm2.zsh ]] || source " + p10kPath + "p10k-iterm2.zsh\n" +
			"  elif [[ $TERM_PROGRAM = \"tmux\" ]]; then\n" +
			"    echo ''; neofetch --bold off\n" +
			"    [[ ! -f " + p10kPath + "p10k-tmux.zsh ]] || source " + p10kPath + "p10k-tmux.zsh\n" +
			"  else\n" +
			"    [[ ! -f " + p10kPath + "p10k-etc.zsh ]] || source " + p10kPath + "p10k-etc.zsh\n" +
			"  fi\n" +
			"else\n" +
			"  [[ ! -f " + p10kPath + "p10k-term.zsh ]] || source " + p10kPath + "p10k-term.zsh\n" +
			"fi\n\n"
		appendContents(profilePath, profileAppend, 0644)
	}

	profileAppend := "# ZSH-COMPLETIONS\n" +
		"if type brew &>/dev/null; then\n" +
		"  FPATH=" + brewPrefix + "share/zsh-completions:$FPATH\n" +
		"  autoload -Uz compinit\n" +
		"  compinit\n" +
		"fi\n\n" +
		"# ZSH SYNTAX HIGHLIGHTING\n" +
		"source " + brewPrefix + "share/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh\n\n" +
		"# ZSH AUTOSUGGESTIONS\n" +
		"source " + brewPrefix + "share/zsh-autosuggestions/zsh-autosuggestions.zsh\n\n" +
		"# Z\n" +
		"source " + brewPrefix + "etc/profile.d/z.sh\n\n" +
		"# ALIAS4SH\n" +
		"source " + homeDir() + "/.config/alias4sh/alias4.sh\n\n" +
		"# Edit\n" +
		"export EDITOR=/usr/bin/vi\n" +
		"edit () { $EDITOR \"$@\" }\n" +
		"#vi () { $EDITOR \"$@\" }\n\n"
	appendContents(profilePath, profileAppend, 0644)

	macLdBar.Stop()
}

func macCLIApp(runOpt string) {
	macLdBar.Suffix = " Installing CLI applications... "
	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install CLI applications!\n"
	macLdBar.Start()

	brewInstall("unzip")
	brewInstall("diffutils")
	brewInstall("transmission-cli")

	if runOpt == "5" || runOpt == "6" || runOpt == "7" {
		brewInstall("curl")
		brewInstall("wget")
		brewInstall("openssh")
		brewInstall("mosh")
		brewInstall("inetutils")
		brewInstall("git")
		brewInstall("git-lfs")
		brewInstall("gh")
		brewInstall("tig")
		brewInstall("exa")
		brewInstall("bat")
		brewInstall("diffr")
		brewInstall("tldr")
		brewInstall("watchman")
		brewInstall("direnv")

		shrcAppend := "# CURL\n" +
			"export PATH=\"" + brewPrefix + "opt/curl/bin:$PATH\"\n" +
			"export LDFLAGS=\"" + brewPrefix + "opt/curl/lib\"\n" +
			"export CPPFLAGS=\"" + brewPrefix + "opt/curl/include\"\n" +
			"export PKG_CONFIG_PATH=\"" + brewPrefix + "opt/curl/lib/pkgconfig\"\n\n" +
			"# DIRENV\n" +
			"eval \"$(direnv hook zsh)\"\n\n"
		appendContents(shrcPath, shrcAppend, 0644)
	}

	if runOpt == "6" || runOpt == "7" {
		brewInstall("make")
		brewInstall("ninja")
		brewInstall("maven")
		brewInstall("gradle")
		brewInstall("rustup-init")
		brewInstall("htop")
		brewInstall("qemu")
		brewInstall("vim")
		brewInstall("neovim")
		brewInstall("curlie")
		brewInstall("jq")
		brewInstall("yq")
		brewInstall("dasel")
		brewInstall("asciinema")
	}

	if runOpt == "7" {
		brewInstall("tor")
		brewInstall("torsocks")
		brewInstall("nmap")
		brewInstall("radare2")
		brewInstall("sleuthkit")
		brewInstall("autopsy")
		brewInstall("virustotal-cli")
	}

	macLdBar.Stop()
}

func macGUIApp(runOpt, adminCode string) {
	macLdBar.Suffix = " Installing GUI applications... "
	macLdBar.Start()

	if runOpt != "7" {
		brewCask("appcleaner", "AppCleaner")
	} else if runOpt == "7" {
		brewCask("sensei", "Sensei")
	}

	brewCask("keka", "Keka")
	brewCask("iina", "IINA")
	brewCask("transmission", "Transmission")
	brewCask("rectangle", "Rectangle")
	brewCask("google-chrome", "Google Chrome")
	brewCask("firefox", "Firefox")
	brewCask("tor-browser", "Tor Browser")
	brewCask("spotify", "Spotify")
	brewCask("signal", "Signal")
	brewCask("discord", "Discord")
	brewCask("slack", "Slack")
	if runOpt == "5" || runOpt == "6" || runOpt == "7" {
		brewCask("jetbrains-space", "JetBrains Space")
	}

	if runOpt == "3" || runOpt == "6" || runOpt == "7" {
		brewCask("dropbox", "Dropbox")
		brewCask("dropbox-capture", "Dropbox Capture")
		brewCask("sketch", "Sketch")
		brewCask("zeplin", "Zeplin")
		brewCask("blender", "Blender")
		brewCask("obs", "OBS")
		brewCaskSudo("loopback", "Loopback", "/Applications/Loopback.app", adminCode)
	}

	if runOpt == "3" || runOpt == "4" || runOpt == "5" {
		brewCaskSudo("blackhole-64ch", "BlackHole (64ch)", "/Library/Audio/Plug-Ins/HAL/BlackHoleXch.driver", adminCode)
	}

	if runOpt == "3" || runOpt == "4" {
		brewCask("eclipse-ide", "Eclipse")
		brewCask("intellij-idea-ce", "IntelliJ IDEA CE")
		brewCask("android-studio", "Android Studio")
		brewCask("visual-studio-code", "Visual Studio Code")
		brewCask("atom", "Atom")
		brewCask("fork", "Fork")
		brewCask("postman", "Postman")
		brewCask("drawio", "draw.io")
		brewCask("httpie", "HTTPie")
		brewCaskSudo("xampp-vm", "xampp-osx-*", "/Applications/Loopback.app", adminCode)
	} else if runOpt == "5" {
		brewCask("iterm2", "iTerm")
		brewCask("intellij-idea", "IntelliJ IDEA")
		brewCask("visual-studio-code", "Visual Studio Code")
		brewCask("atom", "Atom")
		brewCask("neovide", "Neovide")
		brewCask("github", "Github")
		brewCask("fork", "Fork")
		brewCask("docker", "Docker")
		brewCask("tableplus", "TablePlus")
		brewCask("postman", "Postman")
		brewCask("httpie", "HTTPie")
		brewCask("boop", "Boop")
		brewCask("drawio", "draw.io")
	} else if runOpt == "6" || runOpt == "7" {
		brewCask("iterm2", "iTerm")
		brewCask("intellij-idea", "IntelliJ IDEA")
		brewCask("visual-studio-code", "Visual Studio Code")
		brewCask("atom", "Atom")
		brewCask("neovide", "Neovide")
		brewCask("github", "Github")
		brewCask("fork", "Fork")
		brewCask("docker", "Docker")
		brewCaskSudo("vmware-fusion", "VMware Fusion", "/Applications/VMware Fusion.app", adminCode)
		brewCask("tableplus", "TablePlus")
		brewCask("proxyman", "Proxyman")
		brewCask("postman", "Postman")
		brewCask("paw", "Paw")
		brewCask("httpie", "HTTPie")
		brewCask("boop", "Boop")
		brewCask("cmake", "CMake")
		brewCask("drawio", "draw.io")
		brewCask("staruml", "StarUML")
		brewCask("vnc-viewer", "VNC Viewer")
		brewCask("forklift", "ForkLift")
		brewCask("firefox-developer-edition", "Firefox Developer Edition")
	}

	shrcAppend := "# ANDROID STUDIO\n" +
		"export ANDROID_HOME=$HOME/Library/Android/sdk\n" +
		"export PATH=$PATH:$ANDROID_HOME/emulator\n" +
		"export PATH=$PATH:$ANDROID_HOME/tools\n" +
		"export PATH=$PATH:$ANDROID_HOME/tools/bin\n" +
		"export PATH=$PATH:$ANDROID_HOME/platform-tools\n\n"
	appendContents(shrcPath, shrcAppend, 0644)

	if runOpt == "7" {
		brewCask("burp-suite", "Burp Suite Community Edition")
		brewCask("burp-suite-professional", "Burp Suite Professional")
		brewCaskSudo("wireshark", "Wireshark", "/Applications/Wireshark.app", adminCode)
		brewCaskSudo("zenmap", "Zenmap", "/Applications/Zenmap.app", adminCode)
		// Will add Hopper Disassembler
		brewCask("cutter", "Cutter")
		// Will add Ghidra
		brewCask("imazing", "iMazing")
		brewCask("apparency", "Apparency")
		brewCask("suspicious-package", "Suspicious Package")
	}

	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "install GUI applications!\n"
	macLdBar.Stop()
}

func macEnd() {
	macLdBar.Suffix = " Finishing... "
	macLdBar.FinalMSG = lstDot + clrGreen + "Succeed " + clrReset + "clean up homebrew's cache!\n"
	macLdBar.Start()

	shrcAppend := "\n######## ADD CUSTOM VALUES UNDER HERE ########\n\n\n"
	appendContents(shrcPath, shrcAppend, 0644)

	brewUpgrade()
	brewCleanup()
	brewRemoveCache()

	macLdBar.Stop()
}

func macMain(runOpt, runType, brewSts, adminCode string) {
	runEx := lstDot + "Run " + clrPurple + runType + clrReset + " installation\n" +
		lstDot + brewSts + " Homebrew with configure shell"

	if runOpt == "1" {
		fmt.Println(runEx + ".")
		macBegin(adminCode)
		macEnv()
	} else if runOpt == "2" {
		fmt.Println(runEx + ", then install Dependencies, Languages and Terminal/CLI applications " +
			"with set basic preferences.")
		macBegin(adminCode)
		macEnv()
		macDependency(runOpt)
		macLanguage(runOpt, adminCode)
		macTerminal(runOpt)
		macCLIApp(runOpt)
	} else if runOpt == "3" {
		fmt.Println(runEx + ", then install Dependencies, Languages and Terminal/CLI/GUI applications " +
			"with set basic preferences.")
		macBegin(adminCode)
		macEnv()
		macDependency(runOpt)
		macLanguage(runOpt, adminCode)
		macTerminal(runOpt)
		macCLIApp(runOpt)
		macGUIApp(runOpt, adminCode)
	} else if runOpt == "4" {
		fmt.Println(runEx + ", then install Dependencies, Languages and Terminal/CLI/GUI applications " +
			"with set basic preferences.")
		macBegin(adminCode)
		macEnv()
		macDependency(runOpt)
		macLanguage(runOpt, adminCode)
		macServer()
		macDatabase()
		macTerminal(runOpt)
		macCLIApp(runOpt)
		macGUIApp(runOpt, adminCode)
	} else if runOpt == "5" {
		fmt.Println(runEx + ", then install Dependencies, Languages, Server, Database and Terminal/CLI/GUI applications " +
			"with set basic preferences.")
		macBegin(adminCode)
		macEnv()
		macDependency(runOpt)
		macLanguage(runOpt, adminCode)
		macServer()
		macDatabase()
		macTerminal(runOpt)
		macCLIApp(runOpt)
		macGUIApp(runOpt, adminCode)
	} else if runOpt == "6" {
		fmt.Println(runEx + ", then install Dependencies, Languages, Server, Database, management DevTools " +
			"and Terminal/CLI/GUI applications with set basic preferences.")
		macBegin(adminCode)
		macEnv()
		macDependency(runOpt)
		macLanguage(runOpt, adminCode)
		macServer()
		macDatabase()
		macDevVM()
		macTerminal(runOpt)
		macCLIApp(runOpt)
		macGUIApp(runOpt, adminCode)
	} else if runOpt == "7" {
		fmt.Println(runEx + ", then install Dependencies, Languages, Server, Database, management DevTools " +
			"and Terminal/CLI/GUI applications with set basic preferences.")
		macBegin(adminCode)
		macEnv()
		macDependency(runOpt)
		macLanguage(runOpt, adminCode)
		macServer()
		macDatabase()
		macDevVM()
		macTerminal(runOpt)
		macCLIApp(runOpt)
		macGUIApp(runOpt, adminCode)
	}

	macEnd()
}

func macExtend(runOpt, adminCode string) {
	if runOpt != "1" {
		var g4sOpt string
		fmt.Println(clrCyan + "\nConfigure git global easily" + clrReset)
		fmt.Print("Enter [Y] to set git global configuration, or enter any key to exit. ")
		_, errEndOpt := fmt.Scanln(&g4sOpt)
		if errEndOpt != nil {
			g4sOpt = "Enter"
		}
		if g4sOpt == "y" || g4sOpt == "Y" || g4sOpt == "yes" || g4sOpt == "Yes" || g4sOpt == "YES" {
			clearLine(2)
			confG4s()
		} else {
			clearLine(2)
		}
	}

	if runOpt == "3" || runOpt == "6" || runOpt == "7" {
		var rebootOpt string
		fmt.Println(clrCyan + "\nRestart macOS to apply the changes" + clrReset)
		fmt.Print(clrYellow + "ALLOW all application " + clrReset + "on System Preference > Security " +
			"& Privacy > General. (You can restart through the system preference message)\n" +
			"Enter [Y] to restart macOS, or enter any key to exit. ")
		_, errEndOpt := fmt.Scanln(&rebootOpt)
		if errEndOpt != nil {
			rebootOpt = "Enter"
		}
		if rebootOpt == "y" || rebootOpt == "Y" || rebootOpt == "yes" || rebootOpt == "Yes" || rebootOpt == "YES" {
			clearLine(1)
			rebootOS(adminCode)
		} else {
			clearLine(2)
		}
	}
}

func main() {
	var (
		brewSts string
		runOpt  string
		runType string
		endDiv  string
		endMsg  string
	)

	fmt.Println(clrBlue + "\nDev4mac " + clrGrey + "v" + appVer + clrReset + "\n")

	runLdBar.Suffix = " Checking network status... "
	runLdBar.Start()

	if checkExists(cmdPMS) == true {
		brewSts = "Update"
	} else {
		brewSts = "Install"
	}

	if checkNetStatus() != true {
		runLdBar.FinalMSG = clrRed + "Network connect failed" + clrReset + "\n"
		runLdBar.Stop()
		fmt.Println(errors.New(lstDot + "Please check your internet connection.\n"))
		goto exitOpt
	}

	runLdBar.Stop()

	fmt.Println(clrCyan + "The Development tools of Essential and Various for macOS\n" + clrReset +
		lstDot + "Choose an installation option.\n" +
		lstDot + "If you need help, visit https://github.com/leelsey/Dev4os.\n" +
		"\t1. Minimal\n" +
		"\t2. Basic\n" +
		"\t3. Creator\n" +
		"\t4. Beginner\n" +
		"\t5. Developer\n" +
		"\t6. Professional\n" +
		"\t7. Specialist\n" +
		"\t0. Exit\n")

insOpt:
	for {
		fmt.Print("Select command: ")
		_, err := fmt.Scanln(&runOpt)
		if err != nil {
			runOpt = "Null"
		}
		if runOpt == "1" {
			runType = "Minimal"
		} else if runOpt == "2" {
			runType = "Basic"
		} else if runOpt == "3" {
			runType = "Creator"
		} else if runOpt == "4" {
			runType = "Beginner"
		} else if runOpt == "5" {
			runType = "Developer"
		} else if runOpt == "6" {
			runType = "Developer"
		} else if runOpt == "7" {
			runType = "Specialist"
		} else if runOpt == "0" || runOpt == "q" || runOpt == "e" || runOpt == "quit" || runOpt == "exit" {
			fmt.Println(lstDot + "Exited Dev4mac.")
			goto exitOpt
		} else {
			fmt.Println(fmt.Errorf(lstDot + clrYellow + runOpt + clrReset +
				" is invalid option. Please choose number " + clrRed + "0-7" + clrReset + "."))
			tryLoop++
			goto insOpt
		}
		break
	}
	clearLine(12 + tryLoop*2)

	if checkPermission(runOpt, brewSts) == true {
		if adminCode, adminStatus := checkPassword(); adminStatus == true {
			clearLine(1)
			needPermission(adminCode)
			macMain(runOpt, runType, brewSts, adminCode)
			macExtend(runOpt, adminCode)
		} else {
			goto exitOpt
		}
	} else {
		macMain(runOpt, runType, brewSts, "")
		macExtend(runOpt, "")
	}

	endDiv = "\n----------Finished!----------\n"
	endMsg = "Please" + clrRed + " RESTART " + clrReset + "your terminal!\n" +
		lstDot + "Enter this on terminal: source ~/.zprofile && source ~/.zshrc\n" +
		lstDot + "Or restart the Terminal.app by yourself.\n"
	if runOpt == "3" || runOpt == "6" || runOpt == "7" {
		fmt.Println(endDiv + lstDot + "Also you need " + clrRed + "RESTART macOS " + clrReset + " to apply " +
			"the changes after" + clrYellow + "ALLOW all application " + clrReset + "on System Preference " +
			"> Security & Privacy > General\n" + endMsg)
	} else {
		fmt.Println(endDiv + endMsg)
	}

exitOpt:
	return
}
