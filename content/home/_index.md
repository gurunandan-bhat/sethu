+++
title = 'Home'
date = 2024-10-24T10:24:44+05:30
draft = false
[build]
  render = 'never'
[[cascade]]
  [cascade.build]
    render = 'never'
[[cascade]]
  [cascade._target]
  kind = 'page'
  path = '/home/features/**'
  [cascade.params]
  type = 'feature'
[[cascade]]
  [cascade._target]
  kind = 'page'
  path = '/home/reviews/**'
  [cascade.params]
  type = 'review'
[[cascade]]
  [cascade._target]
  kind = 'page'
  path = '/home/footer/**'
  [cascade.params]
  type = 'footer'
+++
