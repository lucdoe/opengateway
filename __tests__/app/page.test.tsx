import {render, screen} from '@testing-library/react'
import {expect, test} from 'vitest'
import Dashboard from '../../app/page'

test('Testing: Home Page', () => {
  render(<Dashboard />)
  expect(
    screen.getByRole('heading', {level: 1, name: 'Gateway Statistics'}),
  ).toBeDefined()

  expect(
    screen.getByRole('heading', {level: 2, name: 'Documentation'}),
  ).toBeDefined()
})
