<!DOCTYPE html>
<html>
  <head>
    <title>PHP Sample</title>
  </head>
  <body>
    <p>Name is <strong><?php echo htmlspecialchars($_POST['name']); ?></strong>.</p>
    <p>Number is <strong><?php echo (int)$_POST['number']; ?></strong>.</p>
    <p>Day is <strong><?php echo $_POST['day']; ?></strong>.</p>
  </body>
</html>
