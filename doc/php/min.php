<!DOCTYPE html>
<html>
  <head>
    <title>PHP Sample</title>
    <style>
      form span {
        font: 15px sans-serif;
        display: inline-block;
        width: 8em;
        text-align: right;
      }
    </style>
  </head>
  <body>
    <form action="action.php" method="post">
      <p><span>Name</span> <input type="text" name="name" /></p>
      <p><span>Number</span> <input type="text" name="number" /></p>
      <p><span>Day</span> <input type="text" name="day" 
        value="<?php echo(date("l", time())); ?>" /></p>
      <p><span>&nbsp;</span> <input type="submit" /></p>
    </form>
  </body>
</html>
