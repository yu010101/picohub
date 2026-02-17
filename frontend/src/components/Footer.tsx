export function Footer() {
  return (
    <footer className="mt-auto border-t border-sand-200 dark:border-sand-800">
      <div className="mx-auto max-w-6xl px-4 py-6">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="font-mono text-xs text-sand-400 dark:text-sand-600">
            picohub &mdash; skill registry for PicoClaw
          </div>
          <div className="flex gap-4 font-mono text-xs text-sand-400 dark:text-sand-600">
            <a
              href="https://github.com/yu01/picohub"
              className="hover:text-sand-700 dark:hover:text-sand-300"
              target="_blank"
              rel="noopener noreferrer"
            >
              src
            </a>
            <span className="text-sand-300 dark:text-sand-700">/</span>
            <a
              href="#"
              className="hover:text-sand-700 dark:hover:text-sand-300"
            >
              docs
            </a>
            <span className="text-sand-300 dark:text-sand-700">/</span>
            <a
              href="#"
              className="hover:text-sand-700 dark:hover:text-sand-300"
            >
              api
            </a>
          </div>
        </div>
      </div>
    </footer>
  );
}
