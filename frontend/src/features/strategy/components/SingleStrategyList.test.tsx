// @vitest-environment jsdom

import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { SingleStrategyList } from './SingleStrategyList';

afterEach(cleanup);

describe('SingleStrategyList plugin discovery', () => {
  it('shows an unknown backend plugin without hardcoded frontend metadata', () => {
    render(
      <SingleStrategyList
        instances={[]}
        availableStrategyNames={['MA', 'MACD']}
        onCreateInstance={vi.fn()}
      />
    );
    fireEvent.click(screen.getByRole('button', { name: /create instance/i }));
    expect(screen.getByRole('option', { name: /MACD/ })).toBeTruthy();
  });
});
