package renderer

import (
	"fmt"
	"strings"
	"testing"
)

// wrapExpected builds the expected DCS passthrough for a Kitty sequence,
// including DECSC/DECRC cursor save/restore.
func wrapExpected(content string) string {
	withSaveRestore := "\x1b7" + content + "\x1b8"
	escaped := strings.ReplaceAll(withSaveRestore, "\x1b", "\x1b\x1b")
	return "\x1bPtmux;" + escaped + "\x1b\\"
}

func TestWrapAllKittySequences_SingleSequence(t *testing.T) {
	r := &TmuxRenderer{inner: NewKittyRenderer()}

	input := "\x1b_Ga=p,i=1,q=2\x1b\\"
	got := r.wrapAllKittySequences(input)
	want := wrapExpected(input)

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWrapAllKittySequences_MultipleChunks(t *testing.T) {
	r := &TmuxRenderer{inner: NewKittyRenderer()}

	chunk1 := "\x1b_Gi=1,f=100,m=1;AAAA\x1b\\"
	chunk2 := "\x1b_Gi=1,m=0;BBBB\x1b\\"
	input := chunk1 + chunk2

	got := r.wrapAllKittySequences(input)
	want := wrapExpected(chunk1) + wrapExpected(chunk2)

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWrapAllKittySequences_PreserveNonKittySequences(t *testing.T) {
	r := &TmuxRenderer{inner: NewKittyRenderer()}

	input := "\x1b[2J" // clear screen — no Kitty sequence
	got := r.wrapAllKittySequences(input)

	if got != input {
		t.Errorf("non-Kitty sequence modified: got %q, want %q", got, input)
	}
}

func TestWrapAllKittySequences_CursorMoveBeforeKitty(t *testing.T) {
	r := &TmuxRenderer{inner: NewKittyRenderer()}

	// Cursor move immediately before Kitty sequence should be pulled into passthrough
	input := "\x1b[H\x1b_Ga=p,i=1,q=2\x1b\\"
	got := r.wrapAllKittySequences(input)

	// \x1b[H with no offset (paneTop=0, paneLeft=0) becomes \x1b[1;1H
	cursorMove := "\x1b[1;1H"
	kittySeq := "\x1b_Ga=p,i=1,q=2\x1b\\"
	want := wrapExpected(cursorMove + kittySeq)

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWrapAllKittySequences_CursorMoveWithPaneOffset(t *testing.T) {
	r := &TmuxRenderer{inner: NewKittyRenderer(), paneTop: 5, paneLeft: 40}

	input := "\x1b[10;20H\x1b_Ga=p,i=1,q=2\x1b\\"
	got := r.wrapAllKittySequences(input)

	// Row 10+5=15, Col 20+40=60
	cursorMove := "\x1b[15;60H"
	kittySeq := "\x1b_Ga=p,i=1,q=2\x1b\\"
	want := wrapExpected(cursorMove + kittySeq)

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWrapAllKittySequences_EmptyString(t *testing.T) {
	r := &TmuxRenderer{inner: NewKittyRenderer()}

	got := r.wrapAllKittySequences("")
	if got != "" {
		t.Errorf("got %q, want empty string", got)
	}
}

func TestWrapAllKittySequences_PlainText(t *testing.T) {
	r := &TmuxRenderer{inner: NewKittyRenderer()}

	input := "hello world"
	got := r.wrapAllKittySequences(input)
	if got != input {
		t.Errorf("got %q, want %q", got, input)
	}
}

func TestFindCursorMoveInPending(t *testing.T) {
	tests := []struct {
		name      string
		paneTop   int
		paneLeft  int
		pending   string
		wantMove  string
		wantFlush int
	}{
		{
			name:      "bare cursor home",
			pending:   "\x1b[H",
			wantMove:  "\x1b[1;1H",
			wantFlush: 0,
		},
		{
			name:      "row and col",
			pending:   "\x1b[10;20H",
			wantMove:  "\x1b[10;20H",
			wantFlush: 0,
		},
		{
			name:      "with pane offset",
			paneTop:   3,
			paneLeft:  50,
			pending:   "\x1b[10;20H",
			wantMove:  "\x1b[13;70H",
			wantFlush: 0,
		},
		{
			name:      "prefix text preserved",
			pending:   "some text\x1b[5;10H",
			wantMove:  "\x1b[5;10H",
			wantFlush: 9, // len("some text")
		},
		{
			name:      "no cursor move",
			pending:   "just text",
			wantMove:  "",
			wantFlush: 9, // len("just text")
		},
		{
			name:      "empty pending",
			pending:   "",
			wantMove:  "",
			wantFlush: 0,
		},
		{
			name:      "row only",
			pending:   "\x1b[5H",
			wantMove:  "\x1b[5;1H",
			wantFlush: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &TmuxRenderer{inner: NewKittyRenderer(), paneTop: tt.paneTop, paneLeft: tt.paneLeft}

			gotMove, gotFlush := r.findCursorMoveInPending(tt.pending)

			if gotMove != tt.wantMove {
				t.Errorf("move: got %q, want %q", gotMove, tt.wantMove)
			}
			if gotFlush != tt.wantFlush {
				t.Errorf("flushEnd: got %d, want %d", gotFlush, tt.wantFlush)
			}
		})
	}
}

func TestWrapAllKittySequences_MixedUploadAndPlacement(t *testing.T) {
	r := &TmuxRenderer{inner: NewKittyRenderer(), paneTop: 2, paneLeft: 10}

	// Simulate: upload chunks + cursor move + placement (like a real frame)
	upload := "\x1b_Gi=1,f=100,a=t,t=d,q=2,m=0;DATA\x1b\\"
	cursorAndPlace := "\x1b[5;3H\x1b_Ga=p,i=1,p=1,c=80,r=24,q=2\x1b\\"
	input := upload + cursorAndPlace

	got := r.wrapAllKittySequences(input)

	// Upload wrapped without cursor move
	wrapUpload := wrapExpected(upload)

	// Placement wrapped with offset cursor move: row 5+2=7, col 3+10=13
	cursorMove := "\x1b[7;13H"
	placeSeq := "\x1b_Ga=p,i=1,p=1,c=80,r=24,q=2\x1b\\"
	wrapPlace := wrapExpected(cursorMove + placeSeq)

	want := wrapUpload + wrapPlace

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestWrapAllKittySequences_DeleteCommand(t *testing.T) {
	r := &TmuxRenderer{inner: NewKittyRenderer()}

	input := fmt.Sprintf("\x1b_Ga=d,d=i,i=%d\x1b\\", 42)
	got := r.wrapAllKittySequences(input)
	want := wrapExpected(input)

	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
