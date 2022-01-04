package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/pingcap/tiup/pkg/utils"
)

type BR struct {
	Path    string
	Version utils.Version
}

type BRBuilder []string

func NewRestore(pdAddr string) *BRBuilder {
	return &BRBuilder{"restore", "full", "-u", pdAddr, "--log-format", "json"}
}

func NewLogRestore(pdAddr string) *BRBuilder {
	return &BRBuilder{"restore", "cdclog", "-u", pdAddr, "--log-format", "json"}
}

func NewBackup(pdAddr string) *BRBuilder {
	return &BRBuilder{"backup", "full", "-u", pdAddr, "--log-format", "json"}
}

func (builder *BRBuilder) Storage(s string) {
	*builder = append(*builder, "-s", s)
}

func (builder *BRBuilder) Build() []string {
	return *builder
}

type BRProcess struct {
	Trace  ProgressTracer
	Handle *exec.Cmd
}

func (br *BR) Execute(ctx context.Context, args ...string) BRProcess {
	cmd := exec.CommandContext(ctx, br.Path, args...)
	r, w := io.Pipe()
	tr := TraceByLog(r)
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	cmd.Env = []string{"AWS_ACCESS_KEY=root", "AWS_SECRET_KEY=a123456;", "BR_LOG_TO_TERM=1"}
	fmt.Println("executing ", args)
	cmd.Start()
	return BRProcess{
		Handle: cmd,
		Trace:  tr,
	}
}

type CdcCtl struct {
	changeFeedId string
	Path         string
	Version      utils.Version
}

type CdcCtlBuilder []string

func NewIncrementalBackup(changeFeedId string, pdAddr string) *CdcCtlBuilder {
	return &CdcCtlBuilder{"cli", "changefeed", "create", "--pd", pdAddr, "--changefeed-id", changeFeedId}
}

func (builder *CdcCtlBuilder) Storage(s string) {
	*builder = append(*builder, "--sink-uri", s)
}

func (builder *CdcCtlBuilder) Build() []string {
	return *builder
}

func GetIncrementalBackup(changeFeedId string, pdAddr string) *CdcCtlBuilder {
	return &CdcCtlBuilder{"cli", "changefeed", "query", "--pd", pdAddr, "--changefeed-id", changeFeedId}
}

func (c *CdcCtl) Execute(ctx context.Context, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, c.Path, args...)
	fmt.Println("executing ", c.Path, args)
	return cmd.Output()
}
