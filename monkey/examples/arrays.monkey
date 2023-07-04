let map = fn(arr, f) {
    let iter = fn(arr, acc) {
        if (len(arr) == 0) {
            acc
        } else {
            iter(tail(arr), push(acc, f(head(arr))));
        }
    };
    iter(arr, []);
};

let reduce = fn(arr, initial, f) {
    let iter = fn(arr, result) {
        if (len(arr) == 0) {
            result
        } else {
            iter(tail(arr), f(result, head(arr)));
        }
    };
    iter(arr, initial);
};

let a = [1, 2, 3, 4];
let double = fn(x) { x * 2 };
let sum = fn(arr) {
    reduce(arr, 0, fn(initial, el) { initial + el });
};
print("map:", map(a, double));
print("sum:", sum(a));