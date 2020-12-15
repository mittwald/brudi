package restic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mittwald/brudi/pkg/cli"
)

const (
	binary = "restic"
)

var (
	ErrRepoAlreadyInitialized = fmt.Errorf("repo already initialized")

	//	createBackupSnapshotIDPattern, createBackupParentSnapshotIDPattern, lsSnapshotSep, lsFileSep                    *regexp.Regexp
	//	forgetSnapshotStartPattern, forgetSnapshotPattern, forgetConcreteSnapshotPattern, forgetSnapshotFinishedPattern *regexp.Regexp

	cmdTimeout = 6 * time.Hour
)

func init() {
	// TODO: use json-log option of restic instead 'regex parsing'-foo
	//	createBackupSnapshotIDPattern = regexp.MustCompile(`snapshot ([0-9a-z]*) saved\n`)
	//	createBackupParentSnapshotIDPattern = regexp.MustCompile(`^using parent snapshot ([0-9a-z]*)\n`)
	//	lsSnapshotSep = regexp.MustCompile(`^snapshot ([0-9a-z]*) of \[(.*)\] at (.*):$`)
	//	lsFileSep = regexp.MustCompile(`^([-rwxd]{10})[ \t]+([0-9a-zA-Z]+)[ \t]+([0-9a-zA-Z]+)[ \t]+([0-9]+)[ \t]+([0-9-]+ [0-9:]+) (.+)$`)

	// forget snapshots
	//	forgetSnapshotStartPattern = regexp.MustCompile(`^remove [0-9]* snapshots:$`)
	//	forgetSnapshotFinishedPattern = regexp.MustCompile(`^[0-9]* snapshots have been removed`)
	//	forgetSnapshotPattern = regexp.MustCompile(`^([0-9a-z]*)[ ].*[0-9]{4}(-[0-9]{2}){2} ([0-9]{2}:){2}[0-9]{2}`)
	//	forgetConcreteSnapshotPattern = regexp.MustCompile(`^removed snapshot ([0-9a-z].*)$`)
}

// InitBackup executes "restic init"
func initBackup(ctx context.Context, globalOpts *GlobalOptions) ([]byte, error) {
	cmd := cli.CommandType{
		Binary:  binary,
		Command: "init",
		Args:    cli.StructToCLI(globalOpts),
	}

	out, err := cli.RunWithTimeout(ctx, cmd, cmdTimeout)
	if err != nil {
		// s3 init-check
		if strings.Contains(string(out), "config already initialized") {
			return out, ErrRepoAlreadyInitialized
		}
		// file init-check
		if strings.Contains(string(out), "file already exists") {
			return out, ErrRepoAlreadyInitialized
		}
		return out, err
	}
	return out, err
}

func parseSnapshotOut(responses Response) (BackupResult, error) {
	var result BackupResult

	var parentSnapshotID string
	var snapshotID string
	for _, resp := range responses.ResponseSummary {
		if resp.SnapshotID != "" {
			snapshotID = resp.SnapshotID
		}
		if resp.ParentSnapshotID != "" {
			parentSnapshotID = resp.SnapshotID
		}
	}
	if parentSnapshotID != "" {
		result.ParentSnapshotID = parentSnapshotID
	}
	if snapshotID == "" {
		return BackupResult{}, fmt.Errorf("failed to parse snapshotID")
	}
	result.SnapshotID = snapshotID
	return result, nil
}

// CreateBackup executes "restic backup" and returns the parent snapshot id (if available) and the snapshot id
func CreateBackup(ctx context.Context, globalOpts *GlobalOptions, backupOpts *BackupOptions, unlock bool) (BackupResult, []byte, error) {
	var out []byte
	var err error

	if unlock {
		unlockOpts := UnlockOptions{
			Flags: &UnlockFlags{
				RemoveAll: false,
			},
		}
		out, err = Unlock(ctx, globalOpts, &unlockOpts)
		if err != nil {
			return BackupResult{}, out, err
		}
	}

	var args []string
	args = cli.StructToCLI(globalOpts)
	args = append(args, cli.StructToCLI(backupOpts)...)

	cmd := newCommand("backup", args...)

	out, err = cli.RunWithTimeout(ctx, cmd, cmdTimeout)
	if err != nil {
		return BackupResult{}, out, err
	}

	backupRes, err := parseSnapshotOut(ReadJSONFromLines(out))
	if err != nil {
		return backupRes, out, err
	}
	return backupRes, nil, nil
}

func ReadJSONFromLines(data []byte) Response {
	r := bytes.NewReader(data)
	var responses Response
	bufReader := bufio.NewReader(r)
	line, _, err := bufReader.ReadLine()
	var content map[string]interface{}
	for err == nil {
		jerr := json.Unmarshal(line, &content)
		if jerr != nil {
			fmt.Errorf("failed to parse response from restic: %s ", line)
			continue
		}
		if content["message_type"] == "status" {
			var resp StatusResponse
			jerr = json.Unmarshal(line, &resp)
			if jerr != nil {
				fmt.Errorf("failed to parse response from restic: %s ", line)
				continue
			}
			responses.ResponseStatus = append(responses.ResponseStatus, resp)
		}
		if content["message_type"] == "summary" {
			var resp SummaryResponse
			jerr = json.Unmarshal(line, &resp)
			if jerr != nil {
				fmt.Errorf("failed to parse response from restic: %s ", line)
				continue
			}
			responses.ResponseSummary = append(responses.ResponseSummary, resp)
		}
		line, _, err = bufReader.ReadLine()
	}
	return responses
}

func newCommand(command string, args ...string) cli.CommandType {
	defaultArgs := []string{"--json"}
	return cli.CommandType{
		Binary:  binary,
		Command: command,
		Args:    append(defaultArgs, args...),
	}
}

// Ls executes "restic ls"
func Ls(ctx context.Context, opts *LsOptions) ([]LsResult, error) {
	var result []LsResult
	cmd := cli.CommandType{
		Binary:  binary,
		Command: "ls",
		Args:    append([]string{"--json"}, cli.StructToCLI(&opts)...),
	}
	out, err := cli.Run(ctx, cmd)
	if err != nil {
		return nil, err
	}

	delimited := strings.Replace(string(out), "\n", ",", -1)

	final := "[" + strings.TrimSuffix(delimited, ",") + "]"
	var current LsResult
	var lsMessages []LsMessage
	err = json.Unmarshal([]byte(final), &lsMessages)
	for _, mess := range lsMessages {
		if mess.ShortID != "" {
			if current.SnapshotID != "" {
				result = append(result, current)
			}
			current = LsResult{
				SnapshotID: mess.ShortID,
				Paths:      mess.Paths,
				Time:       mess.Time,
				Files:      []LsFile{},
			}

			continue
		}

		if !opts.Flags.Long {
			if mess.Type == "file" {
				current.Files = append(current.Files, LsFile{
					Path: mess.Path,
				})
				continue
			}
		}
		if mess.Type == "file" {
			current.Size += mess.Size
			current.Files = append(current.Files, LsFile{
				Permissions: mess.Mode,
				User:        mess.UID,
				Group:       mess.GID,
				Size:        mess.Size,
				Time:        mess.Time,
				Path:        mess.Path,
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
func Forget(
	ctx context.Context, globalOpts *GlobalOptions, forgetOpts *ForgetOptions,
) (
	removedSnapshots []string, output []byte, err error,
) {
	forgetOpts.Flags.Compact = true // make sure compact mode is enabled to parse result correctly

	var args []string
	args = cli.StructToCLI(globalOpts)
	args = append(args, cli.StructToCLI(forgetOpts)...)

	cmd := cli.CommandType{
		Binary:  binary,
		Command: "forget",
		Args:    args,
	}

	out, err := cli.Run(ctx, cmd)
	if err != nil {
		return nil, out, err
	}
	var deletedSnapshots []string
	var forgetResponse ForgetResponse
	err = json.Unmarshal(out, &forgetResponse)
	if err != nil {
		return nil, out, err
	}
	for _, tag := range forgetResponse.tags {
		for _, removed := range tag.Remove {
			deletedSnapshots = append(deletedSnapshots, removed.ID)
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
func Unlock(ctx context.Context, globalOpts *GlobalOptions, unlockOpts *UnlockOptions) ([]byte, error) {
	var args []string
	args = cli.StructToCLI(globalOpts)
	args = append(args, cli.StructToCLI(unlockOpts)...)

	cmd := cli.CommandType{
		Binary:  binary,
		Command: "unlock",
		Args:    args,
	}
	return cli.Run(ctx, cmd)
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
