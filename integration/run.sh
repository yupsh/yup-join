#!/bin/sh
# Integration checks for yup-join, run inside a Debian (GNU coreutils) container.
#
# parity ARGS...  — yup-join FILE1 FILE2 must be byte-identical to GNU `join`.
#
# join requires both inputs sorted on the join field; every fixture below is
# written pre-sorted under LC_ALL=C so GNU `join` and yup-join agree on order.
# Only the implemented surface is exercised: the default join on field 1 and
# the -t field separator. (-1/-2/-a/-e/-o are not implemented; see
# COMPATIBILITY.md.)
set -eu
export LC_ALL=C

fails=0

# Default-separator fixtures (single space, join key is the first field).
printf 'a 1\nb 2\nc 3\n' > /tmp/a.txt
printf 'a x\nb y\nd z\n' > /tmp/b.txt

# Many-to-one fixtures: a repeated key must emit the cross product of groups.
printf 'a 1\na 2\nb 3\n' > /tmp/m1.txt
printf 'a x\nb y\nb z\n' > /tmp/m2.txt

# Tab-separated fixtures for -t.
printf 'a\t1\nb\t2\nc\t3\n' > /tmp/t1.txt
printf 'a\tx\nb\ty\nd\tz\n' > /tmp/t2.txt

# Comma-separated fixtures for -t.
printf 'a,1\nb,2\n' > /tmp/c1.txt
printf 'a,x\nb,y\n' > /tmp/c2.txt

# Bare-key fixtures: a line with no second field must not double the separator.
printf 'a\nb 2\n' > /tmp/k1.txt
printf 'a\nb y\n' > /tmp/k2.txt

parity() {
	ours=$(yup-join "$@" 2>/dev/null || true)
	gnu=$(join "$@" 2>/dev/null || true)
	if [ "$ours" = "$gnu" ]; then
		printf 'ok    parity  join %s\n' "$*"
	else
		printf 'FAIL  parity  join %s\n        gnu:  %s\n        ours: %s\n' "$*" "$gnu" "$ours"
		fails=$((fails + 1))
	fi
}

# Default join on field 1, single-space separator.
parity /tmp/a.txt /tmp/b.txt
# Many-to-one: cross product of equal-key groups.
parity /tmp/m1.txt /tmp/m2.txt
# Bare key on one side: no doubled separator in the output.
parity /tmp/k1.txt /tmp/k2.txt
# -t TAB separator.
parity -t '	' /tmp/t1.txt /tmp/t2.txt
# -t comma separator.
parity -t , /tmp/c1.txt /tmp/c2.txt

if [ "$fails" -ne 0 ]; then
	printf '\n%s check(s) failed\n' "$fails"
	exit 1
fi
printf '\nall checks passed\n'
