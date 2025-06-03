import { NextResponse } from 'next/server';

export async function GET() {
  const config = {
    apiUrl: process.env.API_URL || 'http://localhost:8080/api',
  };

  return NextResponse.json(config);
} 