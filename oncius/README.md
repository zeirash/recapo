# Recapo - Next.js Frontend

A modern order management system for Jastipers (Indonesian cross-border social media sellers), built with Next.js 14, TypeScript, and Material UI.

## 🚀 Features

- **Modern Stack**: Next.js 14 with App Router, TypeScript, and Material UI
- **Mobile-First**: Responsive design optimized for mobile devices
- **Authentication**: JWT-based authentication with React Query
- **Real-time Data**: React Query for efficient data fetching and caching
- **Multi-currency Support**: Built-in support for international transactions
- **Order Management**: Complete CRUD operations for orders, products, and customers

## 🏗️ Project Structure

```
oncius/
├── src/
│   ├── app/                    # Next.js App Router pages
│   │   ├── dashboard/
│   │   ├── orders/
│   │   ├── temp-orders/
│   │   ├── purchase/
│   │   ├── products/
│   │   ├── customers/
│   │   ├── share/[shareToken]/
│   │   ├── login/
│   │   ├── register/
│   │   ├── forgot-password/
│   │   ├── layout.tsx          # Root layout (providers, theme)
│   │   └── page.tsx            # Landing page
│   ├── components/
│   │   ├── layout/             # Layout shell components
│   │   │   ├── index.tsx       # Layout wrapper
│   │   │   ├── Header.tsx      # Auth page header
│   │   │   └── SideMenu.tsx    # Sidebar navigation
│   │   └── ui/                 # Reusable UI primitives
│   │       ├── AddButton.tsx
│   │       ├── CustomerSearchSelect.tsx
│   │       ├── LanguageSwitcher.tsx
│   │       ├── RecapoLogo.tsx
│   │       ├── RecapoLogoText.tsx
│   │       └── SearchInput.tsx
│   ├── constants/
│   ├── hooks/                  # useAuth, useLocale
│   ├── providers/
│   ├── theme/
│   ├── types/
│   └── utils/                  # API client (api.ts)
├── package.json
├── next.config.js
├── tsconfig.json
└── README.md
```

## 🛠️ Tech Stack

- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript
- **Styling**: MUI (Material UI) + Emotion + SASS
- **State Management**: React Query v3 (`react-query`)
- **i18n**: next-intl (EN / ID)
- **Icons**: lucide-react
- **Authentication**: JWT with localStorage
- **Backend**: Golang API on port 4000 (separate service)
- **Database**: PostgreSQL (via Golang backend)

## 📦 Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd recapo/nextjs
   ```

2. **Install dependencies**
   ```bash
   npm install
   # or
   yarn install
   ```

3. **Set up environment variables**
   ```bash
   cp .env.example .env.local
   ```

   Edit `.env.local`:
   ```env
   NEXT_PUBLIC_API_BASE_URL=http://localhost:3000
   ```

4. **Start the development server**
   ```bash
   npm run dev
   # or
   yarn dev
   ```

5. **Open your browser**
   Navigate to [http://localhost:3000](http://localhost:3000)

## 🔧 Development

### Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run start` - Start production server
- `npm run lint` - Run ESLint
- `npm run type-check` - Run TypeScript type checking

### Backend Integration

The Next.js frontend communicates with the Golang backend API. Make sure the backend is running on port 4000:

```bash
# In the arion directory
go run main.go
# or via Docker
docker compose up -d
```

### API Endpoints

The frontend expects the following API endpoints from the Golang backend:

- `POST /login` - User authentication
- `POST /register` - User registration
- `GET /user` - Get current user
- `PATCH /user` - Update user
- `GET /products` - List products
- `POST /products` - Create product
- `PUT /products/:id` - Update product
- `DELETE /products/:id` - Delete product
- `GET /customers` - List customers
- `POST /customers` - Create customer
- `PUT /customers/:id` - Update customer
- `DELETE /customers/:id` - Delete customer
- `GET /orders` - List orders
- `POST /orders` - Create order
- `PUT /orders/:id` - Update order
- `DELETE /orders/:id` - Delete order

## 🎨 Styling

The project uses MUI (Material UI) with a custom theme configured in `src/theme/`. Emotion is used for CSS-in-JS and SASS for global styles in `src/styles/`.

## 🔐 Authentication

Authentication is handled using JWT tokens stored in localStorage. The `useAuth` hook provides:

- Login/logout functionality
- User state management
- Automatic token refresh
- Protected route handling

## 📱 Mobile Optimization

The application is designed with mobile-first principles:

- Touch-friendly buttons (44px minimum)
- Responsive grid layouts
- Optimized typography for mobile screens
- Collapsible navigation menu

## 🚀 Deployment

### Vercel (Recommended)

1. Push your code to GitHub
2. Connect your repository to Vercel
3. Set environment variables in Vercel dashboard
4. Deploy automatically on push

### Docker

```bash
# Build the Docker image
docker build -t recapo-frontend .

# Run the container
docker run -p 3001:3000 recapo-frontend
```


## 📄 License

This project is licensed under the MIT License.

## 🆘 Support

For support and questions:

1. Check the documentation
2. Search existing issues
3. Create a new issue with detailed information

## 🔗 Related Projects

- [Golang Backend (Arion)](../arion/) - REST API service
- [Original Gatsby Frontend](../) - Legacy frontend (deprecated)
