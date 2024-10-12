import * as React from 'react';
import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/')({
  component: HomeComponent,
});

function HomeComponent() {
  return (
    <div className='p-2'>
      <h3>Welcome Home!</h3>
      <h1 className='text-3xl font-bold underline text-red-500'>
        Hello world!
      </h1>
    </div>
  );
}
