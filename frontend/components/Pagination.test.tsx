import React from 'react'
import { render, screen } from '@/lib/test-utils'
import userEvent from '@testing-library/user-event'
import Pagination from './Pagination'

describe('Pagination', () => {
  const mockOnPageChange = jest.fn()

  beforeEach(() => {
    mockOnPageChange.mockClear()
  })

  it('returns null when totalPages is 1 or less', () => {
    const { container } = render(
      <Pagination currentPage={1} totalPages={1} onPageChange={mockOnPageChange} />
    )
    expect(container.firstChild).toBeNull()

    const { container: container0 } = render(
      <Pagination currentPage={1} totalPages={0} onPageChange={mockOnPageChange} />
    )
    expect(container0.firstChild).toBeNull()
  })

  it('renders pagination controls when totalPages > 1', () => {
    render(
      <Pagination currentPage={1} totalPages={5} onPageChange={mockOnPageChange} />
    )
    
    expect(screen.getByText('Previous')).toBeInTheDocument()
    expect(screen.getByText('Next')).toBeInTheDocument()
    expect(screen.getByText('1')).toBeInTheDocument()
    expect(screen.getByText('5')).toBeInTheDocument()
  })

  it('disables Previous button on first page', () => {
    render(
      <Pagination currentPage={1} totalPages={5} onPageChange={mockOnPageChange} />
    )
    
    const previousButton = screen.getByText('Previous')
    expect(previousButton).toBeDisabled()
  })

  it('disables Next button on last page', () => {
    render(
      <Pagination currentPage={5} totalPages={5} onPageChange={mockOnPageChange} />
    )
    
    const nextButton = screen.getByText('Next')
    expect(nextButton).toBeDisabled()
  })

  it('calls onPageChange when clicking page numbers', async () => {
    const user = userEvent.setup()
    render(
      <Pagination currentPage={1} totalPages={5} onPageChange={mockOnPageChange} />
    )
    
    const page3Button = screen.getByText('3')
    await user.click(page3Button)
    
    expect(mockOnPageChange).toHaveBeenCalledWith(3)
  })

  it('highlights the current page', () => {
    render(
      <Pagination currentPage={3} totalPages={5} onPageChange={mockOnPageChange} />
    )
    
    const currentPageButton = screen.getByText('3')
    expect(currentPageButton).toHaveClass('bg-red-600')
    expect(currentPageButton).not.toHaveClass('bg-white')
  })

  it('calls onPageChange with correct value when clicking Previous', async () => {
    const user = userEvent.setup()
    render(
      <Pagination currentPage={3} totalPages={5} onPageChange={mockOnPageChange} />
    )
    
    const previousButton = screen.getByText('Previous')
    await user.click(previousButton)
    
    expect(mockOnPageChange).toHaveBeenCalledWith(2)
  })

  it('calls onPageChange with correct value when clicking Next', async () => {
    const user = userEvent.setup()
    render(
      <Pagination currentPage={3} totalPages={5} onPageChange={mockOnPageChange} />
    )
    
    const nextButton = screen.getByText('Next')
    await user.click(nextButton)
    
    expect(mockOnPageChange).toHaveBeenCalledWith(4)
  })

  describe('Page range calculation', () => {
    it('shows all pages when total pages <= 5', () => {
      render(
        <Pagination currentPage={3} totalPages={5} onPageChange={mockOnPageChange} />
      )
      
      expect(screen.getByText('1')).toBeInTheDocument()
      expect(screen.getByText('2')).toBeInTheDocument()
      expect(screen.getByText('3')).toBeInTheDocument()
      expect(screen.getByText('4')).toBeInTheDocument()
      expect(screen.getByText('5')).toBeInTheDocument()
      expect(screen.queryByText('...')).not.toBeInTheDocument()
    })

    it('shows dots at the end when current page is near the beginning', () => {
      render(
        <Pagination currentPage={2} totalPages={10} onPageChange={mockOnPageChange} />
      )
      
      expect(screen.getByText('1')).toBeInTheDocument()
      expect(screen.getByText('2')).toBeInTheDocument()
      expect(screen.getByText('3')).toBeInTheDocument()
      expect(screen.getByText('4')).toBeInTheDocument()
      expect(screen.getAllByText('...')).toHaveLength(1)
      expect(screen.getByText('10')).toBeInTheDocument()
    })

    it('shows dots at the beginning when current page is near the end', () => {
      render(
        <Pagination currentPage={8} totalPages={10} onPageChange={mockOnPageChange} />
      )
      
      expect(screen.getByText('1')).toBeInTheDocument()
      expect(screen.getAllByText('...')).toHaveLength(1)
      expect(screen.getByText('6')).toBeInTheDocument()
      expect(screen.getByText('7')).toBeInTheDocument()
      expect(screen.getByText('8')).toBeInTheDocument()
      expect(screen.getByText('9')).toBeInTheDocument()
      expect(screen.getByText('10')).toBeInTheDocument()
    })

    it('shows dots on both sides when current page is in the middle', () => {
      render(
        <Pagination currentPage={6} totalPages={12} onPageChange={mockOnPageChange} />
      )
      
      expect(screen.getByText('1')).toBeInTheDocument()
      expect(screen.getAllByText('...')).toHaveLength(2)
      expect(screen.getByText('4')).toBeInTheDocument()
      expect(screen.getByText('5')).toBeInTheDocument()
      expect(screen.getByText('6')).toBeInTheDocument()
      expect(screen.getByText('7')).toBeInTheDocument()
      expect(screen.getByText('8')).toBeInTheDocument()
      expect(screen.getByText('12')).toBeInTheDocument()
    })
  })

  it('has correct dark mode classes', () => {
    render(
      <Pagination currentPage={1} totalPages={5} onPageChange={mockOnPageChange} />
    )
    
    const previousButton = screen.getByText('Previous')
    expect(previousButton).toHaveClass('dark:bg-gray-800')
    expect(previousButton).toHaveClass('dark:border-gray-700')
    expect(previousButton).toHaveClass('dark:text-gray-400')
  })
}) 