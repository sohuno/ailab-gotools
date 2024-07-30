package shellutils

import (
	"bytes"
	"fmt"
	"github.com/sohuno/gotools/stringutils"
	"io"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func RunStringByBachC(cmd string) (string, error) {
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return "", err
	}
	return string(out), err
}

func RunStringWithTimeout(s string, timeoutMs int64) (string, error) {
	buf := bytes.NewBuffer([]byte{})
	cs := stringutils.SplitString(s)
	cmd := CmdFromStrings(cs)
	cmd.Stdout = buf
	cmd.Stderr = buf
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		klog.Errorf("StartCommandFailed Cmd:%s Error:%+v", s, err)
		return "", err
	}
	timer := time.AfterFunc(time.Duration(timeoutMs)*time.Millisecond, func() {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		klog.Warningf("KillProcessGroup:RunStringWithTimeout Command:%s Timeout:%d(ms)", s, timeoutMs)
	})
	err := cmd.Wait()
	timer.Stop()
	return buf.String(), err
}

// Convert a shell command with a series of pipes into
// correspondingly piped list of *exec.Cmd
// If an arg has spaces, this will fail
func RunString(s string) (string, error) {
	buf := bytes.NewBuffer([]byte{})
	sp := strings.Split(s, "|")
	cmds := make([]*exec.Cmd, len(sp))
	// create the commands
	for i, c := range sp {
		cs := strings.Split(strings.TrimSpace(c), " ")
		cmd := CmdFromStrings(cs)
		cmds[i] = cmd
	}

	cmds = AssemblePipes(cmds, nil, buf)
	if err := RunCmds(cmds); err != nil {
		return "", err
	}

	b := buf.Bytes()
	return string(b), nil
}

func GetCmdRunByRoot(cmd string) string {
	curUser, err := user.Current()
	if err != nil {
		klog.Errorf("GetCmdRunByRoot:GetCurrentUserFailed Error:%+v", err)
		return cmd
	}
	if curUser.Username == "admin" {
		if !strings.HasPrefix(cmd, "sudo ") {
			return fmt.Sprintf("sudo %s", cmd)
		}
		return cmd
	} else if curUser.Username == "root" {
		return strings.TrimPrefix(cmd, "sudo ")
	} else {
		return cmd
	}
}

func CmdFromStrings(cs []string) *exec.Cmd {
	if len(cs) == 1 {
		return exec.Command(cs[0])
	} else if len(cs) == 2 {
		return exec.Command(cs[0], cs[1])
	}
	return exec.Command(cs[0], cs[1:]...)
}

// Convert sequence of tokens into commands,
// using "|" as a delimiter
func RunStrings(tokens ...string) (string, error) {
	if len(tokens) == 0 {
		return "", nil
	}
	buf := bytes.NewBuffer([]byte{})
	cmds := []*exec.Cmd{}
	args := []string{}
	// accumulate tokens until a |
	for _, t := range tokens {
		if t != "|" {
			args = append(args, t)
		} else {
			cmds = append(cmds, CmdFromStrings(args))
			args = []string{}
		}
	}
	cmds = append(cmds, CmdFromStrings(args))
	cmds = AssemblePipes(cmds, nil, buf)
	if err := RunCmds(cmds); err != nil {
		return "", fmt.Errorf("%s; %s", err.Error(), string(buf.Bytes()))
	}

	b := buf.Bytes()
	return string(b), nil
}

// Pipe stdout of each command into stdin of next
func AssemblePipes(cmds []*exec.Cmd, stdin io.Reader, stdout io.Writer) []*exec.Cmd {
	cmds[0].Stdin = stdin
	cmds[0].Stderr = stdout
	// assemble pipes
	for i, c := range cmds {
		if i < len(cmds)-1 {
			cmds[i+1].Stdin, _ = c.StdoutPipe()
			cmds[i+1].Stderr = stdout
		} else {
			c.Stdout = stdout
			c.Stderr = stdout
		}
	}
	return cmds
}

// Run series of piped commands
func RunCmds(cmds []*exec.Cmd) error {
	// start processes in descending order
	for i := len(cmds) - 1; i > 0; i-- {
		if err := cmds[i].Start(); err != nil {
			return err
		}
	}
	// run the first process
	if err := cmds[0].Run(); err != nil {
		return err
	}
	// wait on processes in ascending order
	for i := 1; i < len(cmds); i++ {
		if err := cmds[i].Wait(); err != nil {
			return err
		}
	}
	return nil
}

func WithUserAttr(cmd *exec.Cmd, name string) {
	u, err := user.Lookup(name)
	if err != nil {
		klog.Errorf("InvalidUser User:%s", name)
		return
	}
	uid, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		klog.Errorf("ParseUidFailed Uid:%s", u.Uid)
		return
	}
	gid, err := strconv.ParseUint(u.Gid, 10, 32)
	if err != nil {
		klog.Errorf("ParseGidFailed Gid:%s", u.Gid)
		return
	}
	var attr *syscall.SysProcAttr
	if cmd.SysProcAttr != nil {
		attr = cmd.SysProcAttr
	} else {
		attr = &syscall.SysProcAttr{}
	}
	attr.Credential = &syscall.Credential{
		Uid:         uint32(uid),
		Gid:         uint32(gid),
		NoSetGroups: true,
	}
	cmd.SysProcAttr = attr
	cmd.SysProcAttr.Setpgid = true
	klog.V(1).Infof("SetUidGidDone User:%s", name)
}
