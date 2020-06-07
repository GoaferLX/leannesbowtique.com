// Used to toggle the menu on smaller screens when clicking on the menu button
function openNav() {
  var x = document.getElementById("nav");
  if (x.className.indexOf("w3-hide-small") == -1) {
    x.className += " w3-hide-small ";
  } else {
    x.className = x.className.replace(" w3-hide-small ", "");
  }
}
