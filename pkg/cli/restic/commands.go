package restic

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mittwald/brudi/pkg/cli"
	"github.com/mittwald/brudi/pkg/cli/tar"
)

const binary = "restic"

var RepoAlreadyInitialized = fmt.Errorf("repo already initialized")

var createBackupSnapshotIDPattern, createBackupParentSnapshotIDPattern, lsSnapshotSep, lsFileSep *regexp.Regexp
var forgetSnapshotStartPattern, forgetSnapshotPattern, forgetConcreteSnapshotPattern, forgetSnapshotFinishedPattern *regexp.Regexp

func Init() (string, error) {
	createBackupSnapshotIDPattern = regexp.MustCompile(`snapshot ([0-9a-z]*) saved\n`)
	createBackupParentSnapshotIDPattern = regexp.MustCompile(`^using parent snapshot ([0-9a-z]*)\n`)
	lsSnapshotSep = regexp.MustCompile(`^snapshot ([0-9a-z]*) of \[(.*)\] at (.*):$`)
	lsFileSep = regexp.MustCompile(`^([-rwxd]{10})[ \t]+([0-9a-zA-Z]+)[ \t]+([0-9a-zA-Z]+)[ \t]+([0-9]+)[ \t]+([0-9-]+ [0-9:]+) (.+)$`)

	// forget snapshots
	forgetSnapshotStartPattern = regexp.MustCompile(`^remove [0-9]* snapshots:$`)
	forgetSnapshotFinishedPattern = regexp.MustCompile(`^[0-9]* snapshots have been removed`)
	forgetSnapshotPattern = regexp.MustCompile(`^([0-9a-z]*)[ ].*[0-9]{4}(-[0-9]{2}){2} ([0-9]{2}:){2}[0-9]{2}`)
	forgetConcreteSnapshotPattern = regexp.MustCompile(`^removed snapshot ([0-9a-z].*)$`)

	c := &Config{}
	err := c.InitFromViper()
	if err != nil {
		return "", err
	}

	// init restic repo
	var out []byte
	out, err = initBackup()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

// InitBackup executes "restic init"
func initBackup() ([]byte, error) {
	out, err := exec.Command(binary, "init").CombinedOutput()
	if err != nil {
		// s3 init-check
		if strings.Contains(string(out), "config already initialized") {
			return out, RepoAlreadyInitialized
		}
		// file init-check
		if strings.Contains(string(out), "file already exists") {
			return out, RepoAlreadyInitialized
		}
		return out, err
	}
	return out, err
}

func parseSnapshotOut(str string) (BackupResult, error) {
	var result BackupResult

	parentSnapshotID := createBackupParentSnapshotIDPattern.FindStringSubmatch(str)
	snapshotID := createBackupSnapshotIDPattern.FindStringSubmatch(str)
	if len(parentSnapshotID) > 0 {
		result.ParentSnapshotID = parentSnapshotID[1]
	}
	if len(snapshotID) == 0 {
		return BackupResult{}, fmt.Errorf("failed to parse snapshotID: %s ", str)
	}
	result.SnapshotID = snapshotID[1]
	return result, nil
}

// CreateBackup executes "restic backup" and returns the parent snapshot id (if available) and the snapshot id
func CreateBackup(ctx context.Context, opts *BackupOptions, unlock bool) (BackupResult, []byte, error) {
	var out []byte
	var err error

	if unlock {
		unlockOpts := UnlockOptions{
			Flags: &UnlockFlags{
				RemoveAll: false,
			},
		}
		out, err = Unlock(ctx, &unlockOpts)
		if err != nil {
			return BackupResult{}, out, err
		}
	}

	cmd := cli.CommandType{
		Binary:  binary,
		Command: "backup",
		Args:    cli.StructToCLI(opts),
	}
	out, err = cli.RunWithTimeout(ctx, cmd, 6*time.Hour)
	if err != nil {
		return BackupResult{}, out, err
	}

	backupRes, err := parseSnapshotOut(fmt.Sprintf("%s", out))
	if err != nil {
		return backupRes, out, err
	}
	return backupRes, nil, nil
}

// CreateTarBackup executes "restic backup" and returns the parent snapshot id (if available) and the snapshot id
func CreateTarBackup(ctx context.Context, resticOpts *BackupOptions, tarName string) (BackupResult, []byte, error) {
	tarOpts := &tar.Options{
		Flags: &tar.Flags{
			Create:  true,
			File:    "-",
			Exclude: resticOpts.Flags.Exclude,
		},
		Paths: resticOpts.Paths,
	}
	resticOpts.Flags.Stdin = true
	resticOpts.Flags.StdinFilename = tarName
	resticOpts.Flags.Exclude = nil
	resticOpts.Paths = nil

	nice := 19
	ionice := 2
	cmd1 := cli.CommandType{
		Binary: "tar",
		Args:   cli.StructToCLI(tarOpts),
		Nice:   &nice,
		IONice: &ionice,
	}
	cmd2 := cli.CommandType{
		Binary:  binary,
		Command: "backup",
		Args:    cli.StructToCLI(resticOpts),
		Nice:    &nice,
		IONice:  &ionice,
	}
	out, err := cli.RunPipedWithTimeout(ctx, cmd1, cmd2, 6*time.Hour, nil)
	if err != nil {
		return BackupResult{}, out, err
	}

	backupRes, err := parseSnapshotOut(fmt.Sprintf("%s", out))
	if err != nil {
		return backupRes, out, err
	}
	return backupRes, nil, nil
}

// Ls executes "restic ls"
func Ls(ctx context.Context, opts *LsOptions) ([]LsResult, error) {
	var result []LsResult
	cmd := cli.CommandType{
		Binary:  binary,
		Command: "ls",
		Args:    cli.StructToCLI(opts),
	}
	out, err := cli.Run(ctx, cmd)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	var current LsResult

	for _, line := range lines {
		if line == "" {
			continue
		}

		match := lsSnapshotSep.FindStringSubmatch(line)
		if len(match) > 0 {
			if current.SnapshotID != "" {
				result = append(result, current)
			}
			current = LsResult{
				SnapshotID: match[1],
				Paths:      strings.Split(match[2], " "),
				Time:       match[3],
				Files:      []LsFile{},
			}

			continue
		}

		if !opts.Flags.Long {
			current.Files = append(current.Files, LsFile{
				Path: line,
			})
			continue
		}

		fileInfo := lsFileSep.FindStringSubmatch(line)
		if len(fileInfo) > 0 {
			size, err := strconv.ParseUint(fileInfo[4], 10, 64)
			if err != nil {
				continue
			}
			current.Size += size
			current.Files = append(current.Files, LsFile{
				Permissions: fileInfo[1],
				User:        fileInfo[2],
				Group:       fileInfo[3],
				Size:        size,
				Time:        fileInfo[5],
				Path:        fileInfo[6],
			})
		}
	}
	result = append(result, current)

	return result, nil
}

// GetSnapshotSize returns the summed file size of the given snapshots in bytes, based of the "restic ls -l" command.
func GetSnapshotSize(ctx context.Context, snapshotIDs []string) (size uint64) {
	opts := StatsOptions{
		Flags: &StatsFlags{},
		IDs:   snapshotIDs,
	}
	cmd := cli.CommandType{
		Binary:  binary,
		Command: "stats",
		Args:    append([]string{"--json"}, cli.StructToCLI(&opts)...),
	}
	out, err := cli.Run(ctx, cmd)
	if err != nil {
		return
	}

	var stats Stats
	if err = json.Unmarshal(out, &stats); err != nil {
		return
	}
	return stats.TotalSize
}

// GetSnapshotSizeByPath returns the summed file size (filtered by its path)...
// ...of the given snapshot id in bytes, based of the "restic ls -l" command.
func GetSnapshotSizeByPath(ctx context.Context, snapshotID, path string) (size uint64) {
	opts := LsOptions{
		Flags: &LsFlags{
			Long: true,
		},
		SnapshotIDs: []string{snapshotID},
	}
	ls, err := Ls(ctx, &opts)

	if err != nil {
		return
	}

	for _, itm := range ls {
		for _, f := range itm.Files {
			if strings.HasPrefix(f.Path, path) {
				size += f.Size
			}
		}
	}
	return
}

// ListSnapshots executes "restic snapshots"
func ListSnapshots(ctx context.Context, opts *SnapshotOptions) ([]Snapshot, error) {
	cmd := cli.CommandType{
		Binary:  binary,
		Command: "snapshots",
		Args:    append([]string{"--json"}, cli.StructToCLI(opts)...),
	}
	out, err := cli.Run(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var snapshots []Snapshot
	if err := json.Unmarshal(out, &snapshots); err != nil {
		return nil, err
	}
	return snapshots, nil
}

// Find executes "restic find"
func Find(ctx context.Context, opts *FindOptions) ([]FindResult, error) {
	cmd := cli.CommandType{
		Binary:  binary,
		Command: "find",
		Args:    append([]string{"--json"}, cli.StructToCLI(opts)...),
	}
	out, err := cli.Run(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var findResult []FindResult
	if err := json.Unmarshal(out, &findResult); err != nil {
		return nil, err
	}
	return findResult, nil
}

// Check executes "restic check"
func Check(ctx context.Context, flags *CheckFlags) ([]byte, error) {
	cmd := cli.CommandType{
		Binary:  binary,
		Command: "check",
		Args:    cli.StructToCLI(flags),
	}
	return cli.Run(ctx, cmd)
}

// Forget executes "restic forget"
func Forget(ctx context.Context, opts *ForgetOptions) (removedSnapshots []string, output []byte, err error) {
	opts.Flags.Compact = true // make sure compact mode is enabled to parse result correctly
	cmd := cli.CommandType{
		Binary:  binary,
		Command: "forget",
		Args:    cli.StructToCLI(opts),
	}
	out, err := cli.Run(ctx, cmd)
	if err != nil {
		return nil, out, err
	}
	var deletedSnapshots []string

	lines := strings.Split(string(out), "\n")
	start := false
	for _, line := range lines {
		// check if output prints single remove lines (if concrete snapshot id(s) are given)
		concreteID := forgetConcreteSnapshotPattern.FindStringSubmatch(line)
		if len(concreteID) > 0 {
			deletedSnapshots = append(deletedSnapshots, concreteID[1])
			continue
		}

		// check if end of relevant output is reached
		finish := forgetSnapshotFinishedPattern.MatchString(line)
		if finish {
			break
		}

		// check if deleted snapshots block starts
		if !start {
			match := forgetSnapshotStartPattern.MatchString(line)
			if match {
				start = true
			}
			continue
		}

		// check if delete snapshots block ends
		if line == "" {
			start = false
		}

		// check if line contains a deleted snapshot id
		match := forgetSnapshotPattern.FindStringSubmatch(line)
		if len(match) > 0 {
			deletedSnapshots = append(deletedSnapshots, match[1])
		}
	}

	return deletedSnapshots, out, nil
}

// Prune executes "restic prune"
func Prune(ctx context.Context) ([]byte, error) {
	cmd := cli.CommandType{
		Binary:  binary,
		Command: "prune",
		Args:    nil,
	}
	return cli.Run(ctx, cmd)
}

// RebuildIndex executes "restic rebuild-index"
func RebuildIndex(ctx context.Context) ([]byte, error) {
	nice := 19
	ionice := 2
	cmd := cli.CommandType{
		Binary:  binary,
		Command: "rebuild-index",
		Args:    nil,
		Nice:    &nice,
		IONice:  &ionice,
	}
	return cli.Run(ctx, cmd)
}

// RestoreBackup executes "restic restore"
func RestoreBackup(ctx context.Context, opts *RestoreOptions) ([]byte, error) {
	cmd := cli.CommandType{
		Binary:  binary,
		Command: "restore",
		Args:    cli.StructToCLI(opts),
	}
	return cli.Run(ctx, cmd)
}

// Unlock executes "restic unlock"
func Unlock(ctx context.Context, opts *UnlockOptions) ([]byte, error) {
	cmd := cli.CommandType{
		Binary:  binary,
		Command: "unlock",
		Args:    cli.StructToCLI(opts),
	}
	return cli.Run(ctx, cmd)
}

// RestoreTarBackup executes "restic backup" and returns the parent snapshot id (if available) and the snapshot id
func RestoreTarBackup(ctx context.Context, resticOpts *RestoreOptions, tarName string) ([]byte, error) {
	includes := []string{}
	for _, incl := range resticOpts.Flags.Include {
		if !strings.HasPrefix(incl, "/") {
			continue
		}
		includes = append(includes, incl[1:])
	}
	tarOpts := &tar.Options{
		Flags: &tar.Flags{
			Extract: true,
			File:    "-",
			Exclude: resticOpts.Flags.Exclude,
			Target:  resticOpts.Flags.Target,
		},
		Paths: includes,
	}

	dumpOptions := &DumpOptions{
		ID:   resticOpts.ID,
		File: tarName,
	}

	nice := 19
	ionice := 2
	cmd1 := cli.CommandType{
		Binary:  binary,
		Command: "dump",
		Args:    cli.StructToCLI(dumpOptions),
		Nice:    &nice,
		IONice:  &ionice,
	}
	cmd2 := cli.CommandType{
		Binary: "tar",
		Args:   cli.StructToCLI(tarOpts),
		Nice:   &nice,
		IONice: &ionice,
	}
	return cli.RunPipedWithTimeout(ctx, cmd1, cmd2, 6*time.Hour, nil)
}

// Tag executes "restic tag"
func Tag(ctx context.Context, opts *TagOptions) ([]byte, error) {
	cmd := cli.CommandType{
		Binary:  binary,
		Command: "tag",
		Args:    cli.StructToCLI(opts),
	}
	return cli.Run(ctx, cmd)
}
