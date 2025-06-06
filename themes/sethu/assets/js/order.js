import * as params from '@params';

let button = document.getElementById('pay-now');
let title = button.dataset.project;

console.log('Generated with order.js', title);
console.log('Order URL: ', params.orderURL);
