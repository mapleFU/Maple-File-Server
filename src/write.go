package src

/**

call graph:
  sys_open      sysfile.c
    create      sysfile.c
      ialloc    fs.c
      iupdate   fs.c
      dirlink   fs.c
        writei  fs.c
 */
