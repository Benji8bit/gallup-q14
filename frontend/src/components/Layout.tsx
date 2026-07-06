import { useEffect, useState } from 'react';
import { Outlet, Link, useLocation } from 'react-router-dom';
import { Search, Globe } from 'lucide-react';

const NAV_ITEMS = [
  { to: '/survey', label: 'опрос' },
  { to: '/', label: 'о методологии' },
  { to: '/admin', label: 'hr-дашборд' },
];

export const Layout = () => {
  const [scrolled, setScrolled] = useState(false);
  const location = useLocation();
  const isHome = location.pathname === '/';
  const solidHeader = scrolled || !isHome;

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 40);
    onScroll();
    window.addEventListener('scroll', onScroll);
    return () => window.removeEventListener('scroll', onScroll);
  }, []);

  return (
    <div className="min-h-screen flex flex-col bg-page">
      <header className={`toolbar ${solidHeader ? 'toolbar--solid' : 'toolbar--transparent'}`}>
        <div className="toolbar__inner relative">
          <Link to="/" className="flex items-center shrink-0 z-10">
            <img
              src="/favicon.svg"
              alt="Sapiens Solutions"
              className="h-8 sm:h-9 w-auto brightness-0 invert"
            />
          </Link>

          <nav className="toolbar__nav">
            {NAV_ITEMS.map((item) => (
              <Link
                key={item.to}
                to={item.to}
                className={`toolbar__link ${location.pathname === item.to ? 'text-brand-hover' : ''}`}
              >
                {item.label}
              </Link>
            ))}
          </nav>

          <div className="flex items-center gap-4 z-10">
            <div className="toolbar__search">
              <Search className="w-4 h-4 shrink-0" />
              <span className="lowercase text-xs">поиск</span>
            </div>
            <button type="button" className="text-white/80 hover:text-white transition-colors" aria-label="Язык">
              <Globe className="w-5 h-5" />
            </button>
            <div className="flex lg:hidden gap-4">
              <Link to="/survey" className="toolbar__link text-xs">
                опрос
              </Link>
              <Link to="/admin" className="toolbar__link text-xs">
                hr
              </Link>
            </div>
          </div>
        </div>
        {solidHeader && (
          <div className="absolute bottom-0 left-0 right-0 h-px bg-white/15 max-w-[calc(177vmin+10vw)] mx-auto" />
        )}
      </header>

      <main className="flex-1 flex flex-col">
        <Outlet />
      </main>

      <footer className="bg-brand text-white px-6 sm:px-10 lg:px-20 py-12 sm:py-16">
        <div className="max-w-[177vmin] mx-auto flex flex-col md:flex-row justify-between items-start md:items-center gap-8">
          <div>
            <img
              src="/favicon.svg"
              alt="Sapiens Solutions"
              className="h-10 w-auto brightness-0 invert opacity-95"
            />
            <p className="mt-4 text-sm text-white/90 max-w-md leading-relaxed">
              Эксперты в области
              <br />
              аналитических решений
            </p>
          </div>
          <div className="text-sm sm:text-base space-y-2">
            <a href="tel:+74952151757" className="block hover:text-white/80 transition-colors">
              +7 495 215-1757
            </a>
            <a href="mailto:info@sapiens.solutions" className="block hover:text-white/80 transition-colors">
              info@sapiens.solutions
            </a>
            <p className="text-white/40 text-xs mt-4">Опрос проводится полностью анонимно</p>
          </div>
        </div>
      </footer>
    </div>
  );
};
