#!/usr/bin/env python3
"""One-time, opt-in BashCapitalOnline SQLite world reset. Never runs on startup."""
import argparse
import datetime as dt
import hashlib
import os
from pathlib import Path
import sqlite3
import sys

def digest(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()

def active_users(path):
    target = str(path)
    found = []
    for proc in Path("/proc").iterdir():
        if not proc.name.isdigit() or int(proc.name) == os.getpid():
            continue
        try:
            cmd = (proc / "cmdline").read_bytes().replace(b"\0", b" ").decode(errors="replace")
            if ("bashcapital" in cmd.lower() or "server/cmd/server" in cmd.lower()) and target in cmd:
                found.append((proc.name, cmd[:200]))
            for fd in (proc / "fd").iterdir():
                try:
                    if Path(os.readlink(fd)).resolve() == path:
                        found.append((proc.name, "open SQLite file descriptor"))
                        break
                except (OSError, PermissionError):
                    pass
        except (OSError, PermissionError):
            pass
    return found

def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument("--db", required=True, help="Absolute SQLite database path")
    p.add_argument("--backup-dir", required=True, help="Absolute directory outside live database directory")
    p.add_argument("--confirm", help="Must exactly equal RESET-BASHCAPITAL-V2")
    args = p.parse_args()
    db = Path(args.db).expanduser().resolve()
    backup_dir = Path(args.backup_dir).expanduser().resolve()
    if not db.is_absolute() or not backup_dir.is_absolute() or db == backup_dir:
        p.error("invalid paths")
    if not db.is_file() or db.suffix != ".db":
        p.error("database must exist and have .db extension")
    if backup_dir == db.parent or db.parent in backup_dir.parents:
        p.error("backup directory must be outside the live database directory")
    with sqlite3.connect(f"file:{db}?mode=ro", uri=True) as con:
        tables = {row[0] for row in con.execute("SELECT name FROM sqlite_master WHERE type='table'")}
        expected = {"world_save", "accounts", "sessions", "payments"}
        if not expected.issubset(tables):
            p.error("not a recognized BashCapitalOnline database")
        counts = {table: con.execute(f"SELECT COUNT(*) FROM {table}").fetchone()[0] for table in sorted(expected)}
    print("Target:", db)
    print("Rows:", counts)
    print("Action: reset game world, accounts, sessions and payments by replacing the entire game-only SQLite database")
    if args.confirm != "RESET-BASHCAPITAL-V2":
        print("DRY RUN ONLY. Stop the game server and rerun with --confirm RESET-BASHCAPITAL-V2")
        return 0
    if active_users(db):
        p.error("database appears to be in use; stop the server first: " + str(active_users(db)))
    backup_dir.mkdir(parents=True, exist_ok=True)
    stamp = dt.datetime.now(dt.timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    backup = backup_dir / f"{db.stem}-pre-v2-reset-{stamp}.db"
    if backup.exists():
        p.error("backup already exists; retry later")
    # SQLite online-backup API safely captures the complete database, including WAL.
    with sqlite3.connect(f"file:{db}?mode=ro", uri=True) as source, sqlite3.connect(backup) as dest:
        source.backup(dest)
    with sqlite3.connect(backup) as check:
        integrity = check.execute("PRAGMA integrity_check").fetchone()[0]
        backed = {table: check.execute(f"SELECT COUNT(*) FROM {table}").fetchone()[0] for table in sorted(expected)}
    if integrity != "ok" or backed != counts:
        p.error("backup verification failed; original database left unchanged")
    print("Verified backup:", backup, "SHA256:", digest(backup))
    # Replace only the exact selected DB. Refuse WAL/SHM sidecars rather than risking partial state.
    if any(Path(str(db) + suffix).exists() for suffix in ("-wal", "-shm")):
        p.error("SQLite sidecars exist. Stop/checkpoint server and remove sidecars safely before reset; original untouched.")
    archived = db.with_name(db.name + ".pre-v2-reset-" + stamp)
    db.rename(archived)
    try:
        with sqlite3.connect(db) as fresh:
            fresh.execute("PRAGMA user_version=0")
    except Exception:
        if db.exists():
            db.unlink()
        archived.rename(db)
        raise
    db.chmod(0o600)
    print("RESET COMPLETE. Fresh empty DB:", db)
    print("Original also retained at:", archived)
    print("Start the V2 server using this exact -db path; it will create a new V2 world.")
    return 0

if __name__ == "__main__":
    sys.exit(main())
