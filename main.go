package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

func main() {
	if len(os.Args) < 1 {
		log.Println("error executing program.")
	}

	switch os.Args[1] {
	case "run":
		run()
	case "child":
		child()
	default:
		log.Printf("bad command: %s\n", os.Args[1])
	}
}

func run() {
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Clone unix time sharing sistem
	// clone new process id
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
		Unshareflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
	}
	checkError(cmd.Run())
}

func child() {
	fmt.Printf("running: %v as %d\n", os.Args[2:], os.Getpid())

	// cg()
	checkError(syscall.Sethostname([]byte("container")))
	checkError(syscall.Chroot("/go/src/github.com/fn-code/basic-container/rootfs"))
	checkError(syscall.Chdir("/"))
	checkError(syscall.Mount("proc", "proc", "proc", 0, ""))

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	checkError(cmd.Run())
	checkError(syscall.Unmount("proc", 0))
}

func cg() {
	cgroups := "/sys/fs/cgroup/"
	pids := filepath.Join(cgroups, "pids")
	os.Mkdir(filepath.Join(pids, "basic-container"), 0755)

	checkError(ioutil.WriteFile(filepath.Join(pids, "basic-container/pids.max"), []byte("20"), 0700))
	// Removes the new cgroup in place after the container exits
	checkError(ioutil.WriteFile(filepath.Join(pids, "basic-container/notify_on_release"), []byte("1"), 0700))
	checkError(ioutil.WriteFile(filepath.Join(pids, "basic-container/cgroup.procs"), []byte(strconv.Itoa(os.Getpid())), 0700))
}
func checkError(err error) {
	if err != nil {
		log.Println(err)
	}
}
