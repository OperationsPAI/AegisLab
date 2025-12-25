import typer

app = typer.Typer()


@app.command(name="pre-commit")
def run_pre_commit():
    """Run pre-commit checks for Go and Python formatting."""
    from src.git.hook import pre_commit

    pre_commit()


@app.command(name="pre-commit-go")
def run_pre_commit_go():
    """Run pre-commit checks for Go formatting only."""
    from src.git.hook import pre_commit_go

    pre_commit_go()


@app.command(name="pre-commit-python")
def run_pre_commit_python():
    """Run pre-commit checks for Python formatting only."""
    from src.git.hook import pre_commit_python

    pre_commit_python()
