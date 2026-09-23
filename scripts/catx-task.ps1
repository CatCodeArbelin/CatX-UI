param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("status","start","checkpoint","finish")]
    [string]$Action,

    [string]$Branch,
    [string]$Message,
    [switch]$TestsPassed
)

$ErrorActionPreference = "Stop"

function Run-Git {
    param([Parameter(ValueFromRemainingArguments=$true)][string[]]$Args)
    & git @Args
    if ($LASTEXITCODE -ne 0) {
        throw "git $($Args -join ' ') failed with exit code $LASTEXITCODE"
    }
}

function Get-CurrentBranch {
    (& git branch --show-current).Trim()
}

function Assert-Clean {
    $status = (& git status --porcelain)
    if ($status) {
        throw "Working tree is not clean. Commit/stash/review changes first:`n$status"
    }
}

switch ($Action) {
    "status" {
        Write-Host "Branch: $(Get-CurrentBranch)"
        Run-Git status --short --branch
        Run-Git log --oneline --decorate -5
    }

    "start" {
        if (-not $Branch) {
            throw "Use -Branch, e.g. feature/wp-0a-foundation-guardrails"
        }

        Assert-Clean
        Run-Git switch develop
        Run-Git pull --ff-only origin develop

        $exists = (& git branch --list $Branch).Trim()
        if ($exists) {
            Run-Git switch $Branch
            Run-Git merge --ff-only develop
        } else {
            Run-Git switch -c $Branch
        }

        Run-Git push -u origin $Branch
        Write-Host "Started work package on $Branch"
    }

    "checkpoint" {
        if (-not $Message) {
            throw "Use -Message for the checkpoint commit."
        }

        Run-Git add -A

        $staged = (& git diff --cached --name-only)
        if (-not $staged) {
            Write-Host "No staged changes. Nothing to commit."
            exit 0
        }

        Run-Git commit -m $Message
        Run-Git push
        Write-Host "Checkpoint committed and pushed."
    }

    "finish" {
        if (-not $TestsPassed) {
            throw "Refusing to merge. Re-run with -TestsPassed only after required tests actually passed."
        }

        $featureBranch = Get-CurrentBranch
        if ($featureBranch -eq "develop" -or $featureBranch -eq "main") {
            throw "finish must be run from a feature/work-package branch."
        }

        Assert-Clean
        Run-Git push
        Run-Git switch develop
        Run-Git pull --ff-only origin develop
        Run-Git merge --no-ff $featureBranch -m "merge: complete $featureBranch"
        Run-Git push origin develop
        Run-Git branch -d $featureBranch
        Run-Git push origin --delete $featureBranch

        Write-Host "Merged $featureBranch into develop and removed the task branch."
    }
}
