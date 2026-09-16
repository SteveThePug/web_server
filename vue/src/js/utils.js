/**
 * Small shared helpers. Used by components/elle/Elle.vue and views/home/Stamps.vue.
 */

/**
 * Fisher-Yates shuffle. Mutates `array` in place and returns nothing —
 * callers must not do `arr = shuffleArray(arr)`.
 */
export function shuffleArray(array) {
  for (var i = array.length - 1; i > 0; i--) {
    var j = Math.floor(Math.random() * (i + 1));
    var temp = array[i];
    array[i] = array[j];
    array[j] = temp;
  }
}

/** @returns {string} a random opaque colour as a `#RRGGBB` string. */
export function getRandomColor() {
  var letters = "0123456789ABCDEF";
  var color = "#";
  for (var i = 0; i < 6; i++) {
    color += letters[Math.floor(Math.random() * 16)];
  }
  return color;
}
