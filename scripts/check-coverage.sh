#!/usr/bin/env bash
set -eu
profile=${1:-coverage.out}
awk '
NR == 1 { next }
$1 !~ /\/cmd\// {
  total += $2
  if ($3 > 0) covered += $2
}
$1 ~ /\/internal\/domain\// {
  domain_total += $2
  if ($3 > 0) domain_covered += $2
}
END {
  if (total == 0 || domain_total == 0) {
    print "coverage profile must contain non-cmd and domain statements" > "/dev/stderr"
    exit 1
  }
  printf "weighted non-cmd: %d/%d = %.2f%%\n", covered, total, 100 * covered / total
  printf "domain: %d/%d = %.2f%%\n", domain_covered, domain_total, 100 * domain_covered / domain_total
  if (100 * covered < 80 * total || 100 * domain_covered < 90 * domain_total) exit 1
}' "$profile"
