package backends

import (
	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

// EwmhBackend provides workspace switching via EWMH atoms over X11.
// It uses github.com/jezek/xgb (pure Go, no CGo) and gracefully
// returns a nil backend when no X display is available.
type EwmhBackend struct {
	conn                 *xgb.Conn
	root                 xproto.Window
	atomCurrentDesktop   xproto.Atom
	atomNumberOfDesktops xproto.Atom
}

// TryCreate attempts to connect to the X display and intern the EWMH atoms
// required for workspace switching. Returns nil if the X display is not
// reachable or if the atoms cannot be resolved.
func TryCreate() (*EwmhBackend, error) {
	conn, err := xgb.NewConn()
	if err != nil {
		// No X display available — not an error, just a nil backend.
		return nil, nil //nolint:nilerr // graceful degradation when X is not available
	}

	setup := xproto.Setup(conn)
	root := setup.DefaultScreen(conn).Root

	atomCD, err := internAtom(conn, "_NET_CURRENT_DESKTOP")
	if err != nil {
		conn.Close()
		return nil, nil //nolint:nilerr // graceful degradation when atoms cannot be resolved
	}

	atomND, err := internAtom(conn, "_NET_NUMBER_OF_DESKTOPS")
	if err != nil {
		conn.Close()
		return nil, nil //nolint:nilerr // graceful degradation when atoms cannot be resolved
	}

	return &EwmhBackend{
		conn:                 conn,
		root:                 root,
		atomCurrentDesktop:   atomCD,
		atomNumberOfDesktops: atomND,
	}, nil
}

// Switch sets the current desktop to target (0-indexed).
// Returns true if the property was successfully set.
func (b *EwmhBackend) Switch(target int) bool {
	if target < 0 {
		return false
	}
	return setCardinal(b.conn, b.root, b.atomCurrentDesktop, uint32(target)) //nolint:gosec // target is validated non-negative above
}

// GetTotalWorkspaces returns the number of desktops advertised by the
// window manager via _NET_NUMBER_OF_DESKTOPS. Returns 1 on error.
func (b *EwmhBackend) GetTotalWorkspaces() int {
	val, err := getCardinal(b.conn, b.root, b.atomNumberOfDesktops)
	if err != nil {
		return 1
	}
	return int(val)
}

// GetAppendBaseOffset returns the default append offset based on the
// total number of workspaces.
func (b *EwmhBackend) GetAppendBaseOffset() int {
	return AppendBaseOffset(b.GetTotalWorkspaces())
}

func internAtom(conn *xgb.Conn, name string) (xproto.Atom, error) {
	reply, err := xproto.InternAtom(conn, false, uint16(len(name)), name).Reply() //nolint:gosec // Atom names won't exceed 64KB
	if err != nil {
		return 0, err
	}
	return reply.Atom, nil
}

// getCardinal reads a 32-bit cardinal (unsigned int) property from the root
// window. It parses the first 4 bytes of the property value as little-endian.
func getCardinal(conn *xgb.Conn, root xproto.Window, atom xproto.Atom) (uint32, error) {
	reply, err := xproto.GetProperty(conn, false, root, atom,
		xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		return 0, err
	}
	if reply.Format != 32 || len(reply.Value) < 4 {
		return 0, nil
	}
	return xgb.Get32(reply.Value), nil
}

// setCardinal writes a 32-bit cardinal (unsigned int) property on the root
// window. Returns true on success.
func setCardinal(conn *xgb.Conn, root xproto.Window, atom xproto.Atom, value uint32) bool {
	buf := make([]byte, 4)
	xgb.Put32(buf, value)
	err := xproto.ChangePropertyChecked(conn,
		xproto.PropModeReplace, root, atom,
		xproto.AtomCardinal, 32, 1, buf,
	).Check()
	return err == nil
}
