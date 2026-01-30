/**
 * Adds copy buttons to all pre elements that don't already have them.
 * Skips elements already wrapped by ArticleLayout (.code-block) or global (.code-block-global).
 */
export function addCopyButtons(): void {
	document.querySelectorAll('pre').forEach((pre) => {
		const parent = pre.parentElement;

		// Skip if already wrapped
		if (parent?.classList.contains('code-block-global') || parent?.classList.contains('code-block')) {
			return;
		}

		const wrapper = document.createElement('div');
		wrapper.className = 'code-block-global';

		const button = document.createElement('button');
		button.className = 'copy-btn-global';
		button.textContent = 'Copy';
		button.addEventListener('click', async () => {
			const code = pre.querySelector('code')?.textContent || pre.textContent || '';
			await navigator.clipboard.writeText(code);
			button.textContent = 'Copied!';
			setTimeout(() => (button.textContent = 'Copy'), 2000);
		});

		pre.parentNode?.insertBefore(wrapper, pre);
		wrapper.appendChild(button);
		wrapper.appendChild(pre);
	});
}
