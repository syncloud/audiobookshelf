s@    const relpath = req.query.path@    const browserRoot = process.env.FILE_BROWSER_ROOT || '/'\n    let relpath = req.query.path\n    if (relpath \&\& !Path.resolve(relpath).startsWith(browserRoot)) relpath = browserRoot@
s@      directories = await fileUtils.getDirectoriesInPath(relpath || '/', level)@      directories = await fileUtils.getDirectoriesInPath(relpath || browserRoot, level)@
