// Package config abstracts all program configuration.
package config

import (
	"bytes"
	"encoding/json"
	goerr "errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/blang/semver"
	"github.com/micro/go-micro/config"
	"github.com/micro/go-micro/config/encoder"
	jsonenc "github.com/micro/go-micro/config/encoder/json"
	"github.com/micro/go-micro/config/encoder/toml"
	"github.com/micro/go-micro/config/encoder/yaml"
	"github.com/micro/go-micro/config/source"
	"github.com/micro/go-micro/config/source/file"
	"github.com/micro/go-micro/config/source/memory"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

// ErrNoKPVForOS - not all platforms have Kindle Previewer
var ErrNoKPVForOS = errors.New("kindle previewer is not supported on this OS")

//  Internal constants defining if program was invoked via MyHomeLib wrappers.
const (
	MhlNone int = iota
	MhlEpub
	MhlMobi
	MhlUnknown
)

// Logger configuration for single logger.
type Logger struct {
	Level       string `json:"level"`
	Destination string `json:"destination"`
	Mode        string `json:"mode"`
}

// Fb2Mobi provides special support for MyHomeLib.
type Fb2Mobi struct {
	OutputFormat string `json:"output_format"`
	SendToKindle bool   `json:"send_to_kindle"`
}

// Fb2Epub provides special support for MyHomeLib.
type Fb2Epub struct {
	OutputFormat string `json:"output_format"`
}

// SMTPConfig keeps STK configuration.
type SMTPConfig struct {
	DeleteOnSuccess bool   `json:"delete_sent_book"`
	Server          string `json:"smtp_server"`
	Port            int    `json:"smtp_port"`
	User            string `json:"smtp_user"`
	Password        string `json:"smtp_password"`
	From            string `json:"from_mail"`
	To              string `json:"to_mail"`
}

// AuthorName is parsed author name from book metainfo.
type AuthorName struct {
	First  string `json:"first_name"`
	Middle string `json:"middle_name"`
	Last   string `json:"last_name"`
}

func (a *AuthorName) String() string {
	var res string
	if len(a.First) > 0 {
		res = a.First
	}
	if len(a.Middle) > 0 {
		res += " " + a.Middle
	}
	if len(a.Last) > 0 {
		res += " " + a.Last
	}
	return res
}

// MetaInfo keeps book meta-info overwrites from configuration.
type MetaInfo struct {
	ID         string        `json:"id"`
	Title      string        `json:"title"`
	Lang       string        `json:"language"`
	Genres     []string      `json:"genres"`
	Authors    []*AuthorName `json:"authors"`
	SeqName    string        `json:"sequence"`
	SeqNum     int           `json:"sequence_number"`
	Date       string        `json:"date"`
	CoverImage string        `json:"cover_image"`
}

type confMetaOverwrite struct {
	Name string   `json:"name"`
	Meta MetaInfo `json:"meta"`
}

// IsValid checks if we have enough smtp parameters to attempt sending mail.
// It does not attempt actual connection.
func (c *SMTPConfig) IsValid() bool {
	return len(c.Server) > 0 && govalidator.IsDNSName(c.Server) &&
		c.Port > 0 && c.Port <= 65535 &&
		len(c.User) > 0 &&
		len(c.From) > 0 && govalidator.IsEmail(c.From) &&
		len(c.To) > 0 && govalidator.IsEmail(c.To)
}

// Doc format configuration for book processor.
type Doc struct {
	TitleFormat           string  `json:"title_format"`
	AuthorFormat          string  `json:"author_format"`
	AuthorFormatMeta      string  `json:"author_format_meta"`
	AuthorFormatFileName  string  `json:"author_format_file_name"`
	TransliterateMeta     bool    `json:"transliterate_meta"`
	OpenFromCover         bool    `json:"open_from_cover"`
	ChapterPerFile        bool    `json:"chapter_per_file"`
	ChapterLevel          int     `json:"chapter_level"`
	SeqNumPos             int     `json:"series_number_positions"`
	RemovePNGTransparency bool    `json:"remove_png_transparency"`
	ImagesScaleFactor     float64 `json:"images_scale_factor"`
	Stylesheet            string  `json:"style"`
	CharsPerPage          int     `json:"characters_per_page"`
	PagesPerFile          int     `json:"pages_per_file"`
	Hyphenate             bool    `json:"insert_soft_hyphen"`
	NoNBSP                bool    `json:"ignore_nonbreakable_space"`
	UseBrokenImages       bool    `json:"use_broken_images"`
	FileNameFormat        string  `json:"file_name_format"`
	FileNameTransliterate bool    `json:"file_name_transliterate"`
	FixZip                bool    `json:"fix_zip_format"`
	//
	DropCaps struct {
		Create        bool   `json:"create"`
		IgnoreSymbols string `json:"ignore_symbols"`
	} `json:"dropcaps"`
	Notes struct {
		BodyNames []string `json:"body_names"`
		Mode      string   `json:"mode"`
	} `json:"notes"`
	Annotation struct {
		Create   bool   `json:"create"`
		AddToToc bool   `json:"add_to_toc"`
		Title    string `json:"title"`
	} `json:"annotation"`
	TOC struct {
		Type              string `json:"type"`
		Title             string `json:"page_title"`
		Placement         string `json:"page_placement"`
		MaxLevel          int    `json:"page_maxlevel"`
		NoTitleChapters   bool   `json:"include_chapters_without_title"`
		BookTitleFromMeta bool   `json:"book_title_from_meta"`
	} `json:"toc"`
	Cover struct {
		Default   bool   `json:"default"`
		ImagePath string `json:"image_path"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		Placement string `json:"stamp_placement"`
		Font      string `json:"stamp_font"`
	} `json:"cover"`
	Vignettes struct {
		Create bool                         `json:"create"`
		Images map[string]map[string]string `json:"images"`
	} `json:"vignettes"`
	//
	Transformations map[string]map[string]string `json:"transform"`
	//
	Kindlegen struct {
		Path             string `json:"path"`
		CompressionLevel int    `json:"compression_level"`
		Verbose          bool   `json:"verbose"`
		NoOptimization   bool   `json:"no_mobi_optimization"`
		RemovePersonal   bool   `json:"remove_personal_label"`
		PageMap          string `json:"generate_apnx"`
		ForceASIN        bool   `json:"force_asin_on_azw3"`
	} `json:"kindlegen"`
	//
	KPreViewer struct {
		Path string `json:"path"`
	} `json:"kindle_previewer"`
}

// names of supported vignettes
const (
	VigBeforeTitle = "before_title"
	VigAfterTitle  = "after_title"
	VigChapterEnd  = "chapter_end"
)

// Config keeps all configuration values.
type Config struct {
	// Internal implementation - keep it local, could be replaced
	Path string
	cfg  config.Config

	// Actual configuration used everywhere - immutable
	ConsoleLogger Logger
	FileLogger    Logger
	Doc           Doc
	SMTPConfig    SMTPConfig
	Fb2Mobi       Fb2Mobi
	Fb2Epub       Fb2Epub
	Overwrites    map[string]MetaInfo
}

var defaultConfig = []byte(`{
  "document": {
    "title_format": "{(#ABBRseries{ - #padnumber}) }#title",
    "author_format": "#l{ #f}{ #m}",
    "chapter_per_file": true,
    "chapter_level": 2147483647,
    "series_number_positions": 2,
    "characters_per_page": 2300,
    "pages_per_file": 2147483647,
    "fix_zip_format": true,
    "dropcaps": {
      "ignore_symbols": "'\"-.…0123456789‒–—«»“”\u003c\u003e"
    },
    "vignettes": {
      "create": true,
      "images": {
        "default": {
          "after_title": "profiles/vignettes/title_after.png",
          "before_title": "profiles/vignettes/title_before.png",
          "chapter_end": "profiles/vignettes/chapter_end.png"
        },
        "h0": {
          "after_title": "none",
          "before_title": "none",
          "chapter_end": "none"
        }
      }
    },
    "kindlegen": {
      "compression_level": 1,
      "remove_personal_label": true,
      "generate_apnx": "none"
    },
    "cover": {
      "height": 1680,
      "width": 1264
    },
    "notes": {
      "body_names": [ "notes", "comments" ],
      "mode": "default"
    },
    "annotation": {
      "title": "Annotation"
    },
    "toc": {
      "type": "normal",
      "page_title": "Content",
      "page_placement": "after",
      "page_maxlevel": 2147483647
    }
  },
  "logger": {
    "console": {
      "level": "normal"
    },
    "file": {
      "destination": "conversion.log",
      "level": "debug",
      "mode": "append"
    }
  },
  "fb2mobi": {
    "output_format": "mobi"
  },
  "fb2epub": {
    "output_format": "epub"
  }
}`)

// BuildConfig loads configuration from file.
func BuildConfig(fname string) (*Config, error) {

	var base string

	c := config.NewConfig()
	switch {
	case fname == "-":
		// stdin - json format ONLY
		source, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return nil, errors.Wrap(err, "unable to read configuration from stdin")
		}
		err = c.Load(
			// default values
			memory.NewSource(memory.WithJSON(defaultConfig)),
			// overwrite
			memory.NewSource(memory.WithJSON(source)))

		if err != nil {
			return nil, errors.Wrap(err, "unable to read configuration from stdin")
		}
		if base, err = os.Getwd(); err != nil {
			return nil, errors.Wrap(err, "unable to get working directory")
		}
	case len(fname) > 0:
		// from file
		var enc encoder.Encoder
		switch strings.ToLower(filepath.Ext(fname)) {
		case ".yml":
			fallthrough
		case ".yaml":
			enc = yaml.NewEncoder()
		case ".toml":
			enc = toml.NewEncoder()
		default:
			enc = jsonenc.NewEncoder()
		}
		err := c.Load(
			// default values
			memory.NewSource(memory.WithJSON(defaultConfig)),
			// overwrite
			file.NewSource(file.WithPath(fname), source.WithEncoder(enc)))

		if err != nil {
			return nil, errors.Wrapf(err, "unable to read configuration file (%s)", fname)
		}
		if base, err = filepath.Abs(filepath.Dir(fname)); err != nil {
			return nil, errors.Wrap(err, "unable to get configuration directory")
		}
	default:
		// default values
		err := c.Load(memory.NewSource(memory.WithJSON(defaultConfig)))
		if err != nil {
			return nil, errors.Wrap(err, "unable to prepare default configuration")
		}
	}

	conf := Config{cfg: c, Path: base, Overwrites: make(map[string]MetaInfo)}
	if err := c.Get("logger", "console").Scan(&conf.ConsoleLogger); err != nil {
		return nil, errors.Wrap(err, "unable to read console logger configuration")
	}
	if err := c.Get("logger", "file").Scan(&conf.FileLogger); err != nil {
		return nil, errors.Wrap(err, "unable to read file logger configuration")
	}
	if err := c.Get("document").Scan(&conf.Doc); err != nil {
		return nil, errors.Wrap(err, "unable to read document format configuration")
	}
	if err := c.Get("fb2mobi").Scan(&conf.Fb2Mobi); err != nil {
		return nil, errors.Wrap(err, "unable to read fb2mobi cnfiguration")
	}
	if err := c.Get("fb2epub").Scan(&conf.Fb2Epub); err != nil {
		return nil, errors.Wrap(err, "unable to read fb2epub cnfiguration")
	}
	if err := c.Get("sendtokindle").Scan(&conf.SMTPConfig); err != nil {
		return nil, errors.Wrap(err, "unable to read send to kindle cnfiguration")
	}

	var metas []confMetaOverwrite
	if err := c.Get("overwrites").Scan(&metas); err != nil {
		return nil, errors.Wrap(err, "unable to read meta information overwrites")
	}
	for _, meta := range metas {
		name := filepath.ToSlash(meta.Name)
		if _, exists := conf.Overwrites[name]; !exists {
			conf.Overwrites[name] = meta.Meta
		}
	}

	// some defaults
	if conf.Doc.Kindlegen.CompressionLevel < 0 || conf.Doc.Kindlegen.CompressionLevel > 2 {
		conf.Doc.Kindlegen.CompressionLevel = 1
	}
	// to keep old behavior
	if len(conf.Doc.AuthorFormatMeta) == 0 {
		conf.Doc.AuthorFormatMeta = conf.Doc.AuthorFormat
	}
	if len(conf.Doc.AuthorFormatFileName) == 0 {
		conf.Doc.AuthorFormatFileName = conf.Doc.AuthorFormat
	}
	return &conf, nil
}

// GetBytes returns configuration the way it was read from various sources, before unmarshaling.
func (conf *Config) GetBytes() ([]byte, error) {
	// do some pretty-printing
	var out bytes.Buffer
	err := json.Indent(&out, conf.cfg.Bytes(), "", "  ")
	return out.Bytes(), err
}

// Transformation is used to specify additional text processsing during conversion.
type Transformation struct {
	From string
	To   string
}

// GetTransformation returns pointer to named text transformation of nil if none eavailable.
func (conf *Config) GetTransformation(name string) *Transformation {

	if len(conf.Doc.Transformations) == 0 {
		return nil
	}

	m, exists := conf.Doc.Transformations[name]
	if !exists {
		return nil
	}

	if f, ok := m["from"]; ok && len(f) > 0 {
		return &Transformation{
			From: f,
			To:   m["to"],
		}
	}
	return nil
}

// GetOverwrite returns pointer to information to be used instead of parsed data.
func (conf *Config) GetOverwrite(name string) *MetaInfo {

	if len(conf.Overwrites) == 0 {
		return nil
	}

	// start from most specific

	// NOTE: all path separators were converted to slash before being added to map
	name = filepath.ToSlash(name)
	for {
		if i, ok := conf.Overwrites[name]; ok {
			return &i
		}
		parts := strings.SplitN(name, "/", 1)
		if len(parts) <= 1 {
			break
		}
		name = parts[1]
	}

	// not found - see if we have generic overwrite
	name = "*"
	if i, ok := conf.Overwrites[name]; ok {
		return &i
	}
	return nil
}

// GetKindlegenPath provides platform specific path to the kindlegen executable.
func (conf *Config) GetKindlegenPath() (string, error) {

	fname := conf.Doc.Kindlegen.Path
	expath, err := os.Executable()
	if err != nil {
		return "", errors.Wrap(err, "unable to detect program path")
	}
	if expath, err = filepath.Abs(filepath.Dir(expath)); err != nil {
		return "", errors.Wrap(err, "unable to calculate program path")
	}

	if len(fname) > 0 {
		if !filepath.IsAbs(fname) {
			fname = filepath.Join(expath, fname)
		}
	} else {
		fname = filepath.Join(expath, kindlegen())
	}
	if _, err = os.Stat(fname); err != nil {
		return "", errors.Wrap(err, "unable to find kindlegen")
	}
	return fname, nil
}

var (
	reKPVver           = regexp.MustCompile(`^Kindle Previewer ([0-9]+\.[0-9]+\.[0-9]+) Copyright (c) Amazon.com$`)
	minSupportedKPVver = semver.Version{Major: 3, Minor: 32, Patch: 0}
)

// GetKPVPath provides platform specific path to the kindle previever executable.
func (conf *Config) GetKPVPath() (string, error) {

	var err error

	kpath := conf.Doc.KPreViewer.Path
	if len(kpath) > 0 {
		if !filepath.IsAbs(kpath) {
			return "", errors.Errorf("path to kindle previewer must be absolute path: %s", kpath)
		}
	} else {
		kpath, err = kpv()
		if err != nil {
			return "", errors.Wrap(err, "problem getting kindle previewer path")
		}
	}
	if _, err := os.Stat(kpath); err != nil {
		return "", errors.Wrapf(err, "unable to find kindle previewer: %s", kpath)
	}

	var out []byte
	if out, err = exec.Command(kpath, "-help").CombinedOutput(); err != nil {
		return "", errors.Wrapf(err, "unable to run kindle previewer: %s", kpath)
	}
	res := reKPVver.FindSubmatch(out)
	if len(res) < 2 {
		return "", errors.New("unable to find kindle previewer version")
	}
	var ver semver.Version
	if ver, err = semver.Parse(string(res[1])); err != nil {
		return "", errors.Wrap(err, "unable to parse kindle previewer version")
	}
	if minSupportedKPVver.GT(ver) {
		errors.Errorf("unsupported version %s of kindle previewer is installed (required version %s or newer)", ver, minSupportedKPVver)
	}
	return kpath, nil
}

// GetActualBytes returns actual configuration, including fields initialized by default.
func (conf *Config) GetActualBytes() ([]byte, error) {

	// For convinience create temporary configuration structure with actual values
	a := struct {
		B struct {
			Cl Logger `json:"console"`
			Fl Logger `json:"file"`
		} `json:"logger"`
		D Doc        `json:"document"`
		E SMTPConfig `json:"sendtokindle"`
		F Fb2Mobi    `json:"fb2mobi"`
		G Fb2Epub    `json:"fb2epub"`
		H []struct {
			Name string   `json:"name"`
			Meta MetaInfo `json:"meta"`
		} `json:"overwrites"`
	}{}
	a.B.Cl = conf.ConsoleLogger
	a.B.Fl = conf.FileLogger
	a.D = conf.Doc
	a.E = conf.SMTPConfig
	a.F = conf.Fb2Mobi
	a.G = conf.Fb2Epub

	for k, v := range conf.Overwrites {
		s := struct {
			Name string   `json:"name"`
			Meta MetaInfo `json:"meta"`
		}{Name: filepath.FromSlash(k), Meta: v}
		a.H = append(a.H, s)
	}

	// Marshall it to json
	b, err := json.Marshal(a)
	if err != nil {
		return []byte{}, err
	}

	// And pretty-print it
	var out bytes.Buffer
	err = json.Indent(&out, b, "", "  ")
	return out.Bytes(), err
}

// PrepareLog returns our standard logger. It prepares zap logger for use by the program.
func (conf *Config) PrepareLog() (*zap.Logger, error) {

	// Console - split stdout and stderr, handle colors and redirection

	ec := zap.NewDevelopmentEncoderConfig()
	ec.EncodeCaller = nil
	if EnableColorOutput(os.Stdout) {
		ec.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		ec.EncodeLevel = zapcore.CapitalLevelEncoder
	}
	consoleEncoderLP := zapcore.NewConsoleEncoder(ec)

	ec = zap.NewDevelopmentEncoderConfig()
	ec.EncodeCaller = nil
	if EnableColorOutput(os.Stderr) {
		ec.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		ec.EncodeLevel = zapcore.CapitalLevelEncoder
	}
	consoleEncoderHP := newEncoder(ec) // filter errorVerbose

	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})

	var consoleCoreHP, consoleCoreLP zapcore.Core
	switch conf.ConsoleLogger.Level {
	case "normal":
		consoleCoreLP = zapcore.NewCore(consoleEncoderLP, zapcore.Lock(os.Stdout),
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return zapcore.InfoLevel <= lvl && lvl < zapcore.ErrorLevel
			}))
		consoleCoreHP = zapcore.NewCore(consoleEncoderHP, zapcore.Lock(os.Stderr), highPriority)
	case "debug":
		consoleCoreLP = zapcore.NewCore(consoleEncoderLP, zapcore.Lock(os.Stdout),
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
				return zapcore.DebugLevel <= lvl && lvl < zapcore.ErrorLevel
			}))
		consoleCoreHP = zapcore.NewCore(consoleEncoderHP, zapcore.Lock(os.Stderr), highPriority)
	default:
		consoleCoreLP = zapcore.NewNopCore()
		consoleCoreHP = zapcore.NewNopCore()
	}

	// File

	opener := func(fname, mode string) (f *os.File, err error) {
		flags := os.O_CREATE | os.O_WRONLY
		if mode == "append" {
			flags |= os.O_APPEND
		} else {
			flags |= os.O_TRUNC
		}
		if f, err = os.OpenFile(fname, flags, 0644); err != nil {
			return nil, err
		}
		return f, nil
	}

	var (
		fileEncoder  zapcore.Encoder
		fileCore     zapcore.Core
		logLevel     zap.AtomicLevel
		logRequested bool
	)
	switch conf.FileLogger.Level {
	case "debug":
		fileEncoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		logLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
		logRequested = true
	case "normal":
		fileEncoder = zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		logLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
		logRequested = true
	}

	if logRequested {
		if f, err := opener(conf.FileLogger.Destination, conf.FileLogger.Mode); err == nil {
			fileCore = zapcore.NewCore(fileEncoder, zapcore.Lock(f), logLevel)
		} else {
			return nil, errors.Wrapf(err, "unable to access file log destination (%s)", conf.FileLogger.Destination)
		}
	} else {
		fileCore = zapcore.NewNopCore()
	}

	return zap.New(zapcore.NewTee(consoleCoreHP, consoleCoreLP, fileCore), zap.AddCaller()), nil
}

// When logging error to console - do not output verbose message.

type consoleEnc struct {
	zapcore.Encoder
}

func newEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	return consoleEnc{zapcore.NewConsoleEncoder(cfg)}
}

func (c consoleEnc) Clone() zapcore.Encoder {
	return consoleEnc{c.Encoder.Clone()}
}

func (c consoleEnc) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	var newFields []zapcore.Field
	for _, f := range fields {
		if f.Type == zapcore.ErrorType {
			e := f.Interface.(error)
			f.Interface = goerr.New(e.Error())
		}
		newFields = append(newFields, f)
	}
	return c.Encoder.EncodeEntry(ent, newFields)
}
