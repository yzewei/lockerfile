package gockerfile

import (
	"github.com/moby/buildkit/client/llb"
	"github.com/moby/buildkit/util/system"
	"github.com/po3rin/gockerfile/config"
)

// Gocker2LLB converts gockerfile yml to LLB.
func Gocker2LLB(c *config.Config) (llb.State, *Image, error) {
	state := buildkit(c)
	imgCfg := NewImageConfig()
	return state, imgCfg, nil
}

func goBuildBase() llb.State {
	goAlpine := llb.Image("yangzewei2023/golang:1.21-alpine")
	return goAlpine.
		AddEnv("PATH", "/usr/local/go/bin:"+system.DefaultPathEnv("linux")).
		AddEnv("GO111MODULE", "on").
		Run(llb.Shlex("apk add --no-cache git")).
		Root()
}

func goRepo(s llb.State, repo, ref string, g ...llb.GitOption) func(ro ...llb.RunOption) llb.State {
	dir := "/go/src/" + repo
	return func(ro ...llb.RunOption) llb.State {
		es := s.Dir(dir).Run(ro...)
		es.AddMount(dir, llb.Git(repo, ref, g...))
		return es.AddMount(dir+"/bin", llb.Scratch())
	}
}

func buildkit(c *config.Config) llb.State {
	builder := goRepo(goBuildBase(), c.Repo, c.Ver)
	built := builder(llb.Shlex("go build -o ./bin/server " + c.Path))
	r := llb.Image("yangzewei2023/alpine:v3.18-base").With(
		copyAll(built, "/bin"),
	)
	return r
}

func copyAll(src llb.State, destPath string) llb.StateOption {
	return copyFrom(src, "/.", destPath)
}

// copyFrom has similar semantics as `COPY --from`
func copyFrom(src llb.State, srcPath, destPath string) llb.StateOption {
	return func(s llb.State) llb.State {
		return copy(src, srcPath, s, destPath)
	}
}

func copy(src llb.State, srcPath string, dest llb.State, destPath string) llb.State {
	cpImage := llb.Image("yangzewei2023/alpine:v3.18-base")
	cp := cpImage.Run(llb.Shlexf("cp -a /src%s /dest%s", srcPath, destPath))
	cp.AddMount("/src", src)
	return cp.AddMount("/dest", dest)
}
