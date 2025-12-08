import typer

app = typer.Typer()


@app.command(name="pre-commit")
def run_pre_commit():
    """Run pre-commit checks for Go and Python formatting."""
    from src.git.hook import pre_commit

    pre_commit()
