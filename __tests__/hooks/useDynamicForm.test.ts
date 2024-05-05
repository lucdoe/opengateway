import {act, renderHook} from '@testing-library/react'
import {ChangeEvent} from 'react'
import {test} from 'vitest'
import {Field, useDynamicForm} from '../../hooks/useDynamicForm'

test('useDynamicForm: updates form data correctly', ({expect}) => {
  const initialFields: Field[] = [
    {name: 'test', fieldtype: 'text', label: 'Test Label'},
  ]

  const {result} = renderHook(() => useDynamicForm(initialFields))

  act(() => {
    const event = {
      target: {value: 'new value', type: 'text'},
    } as ChangeEvent<HTMLTextAreaElement>

    return result.current.handleInputChange('test', event)
  })

  expect(result.current.formData).toEqual({test: 'new value'})
})

test('useDynamicForm: handles checkbox input correctly', ({expect}) => {
  const initialFields: Field[] = [
    {name: 'test', fieldtype: 'text', label: 'Test Label'},
  ]
  const {result} = renderHook(() => useDynamicForm(initialFields))

  act(() => {
    result.current.handleInputChange('test', {
      target: {checked: true, type: 'checkbox'},
    } as ChangeEvent<HTMLInputElement>)
  })

  expect(result.current.formData).toEqual({test: true})
})

test('useDynamicForm: handles boolean input correctly', ({expect}) => {
  const initialFields: Field[] = [
    {name: 'test', fieldtype: 'text', label: 'Test Label'},
  ]
  const {result} = renderHook(() => useDynamicForm(initialFields))

  act(() => {
    result.current.handleInputChange('test', true)
  })

  expect(result.current.formData).toEqual({test: true})
})
