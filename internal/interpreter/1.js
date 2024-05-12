function a(i) { if (i == 0) "Exit"; else { print(i); return a(i - 1); } } pprint("exit1", a(3));
function a(i) { for (; i > 0; i--) { print(i); if (i == 0) return 1; else if (i == 1) return 2; }; } pprint("exit1", a(3));


function a(i) { for (; i > 0; i = i - 1) { print(i); if (i == 0) { return 1; } else if (i == 1) { return 2; } } } pprint("exit1", a(3));