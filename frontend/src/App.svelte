<script>
	let service = "";
	let repoUrl = "";
	let group_path = "";
	let username = "";
	let password = "@";
	let token = "";
	let repos = [];

	async function api(url = "", data = {}, params = {}) {
		const response = await fetch(url, { ...params });
		return response.json();
	}

	function extractServiceAndGroupPath(url) {
		if (url.includes("github.com")) {
			service = "github";
			const parts = url.split("github.com/");
			if (parts.length > 1) {
				group_path = parts[1].replace(/\.git$/, "");
			} else {
				alert("URL inválida para GitHub");
				return false;
			}
		} else if (url.includes("gitlab.com")) {
			service = "gitlab";
			const parts = url.split("gitlab.com/");
			if (parts.length > 1) {
				group_path = parts[1].replace(/\.git$/, "");
			} else {
				alert("URL inválida para GitLab");
				return false;
			}
		} else {
			alert("Serviço não suportado");
			return false;
		}
		return true;
	}

	async function listRepos() {
		if (!extractServiceAndGroupPath(repoUrl)) {
			return;
		}

		api(
			"http://localhost:8080/list",
			{},
			{
				method: "POST",
				headers: {
					"Content-Type": "application/x-www-form-urlencoded",
				},
				body: new URLSearchParams({
					service: service,
					group_path: group_path,
					username: username,
					token: token,
				}),
			},
		).then((data) => {
			console.log(data);
			repos = data.repos;
		});
	}

	async function cloneRepo() {
		const response = await fetch("http://localhost:8080/clone", {
			method: "POST",
			headers: {
				"Content-Type": "application/x-www-form-urlencoded",
			},
			body: new URLSearchParams({
				repo_url: repoUrl,
				username: username,
				password: password,
			}),
		});
		const result = await response.text();
		alert(result);
	}
</script>

<main>
	<h1>Clone GIT Repository</h1>
	<form on:submit|preventDefault={listRepos}>
		Service:
		<select bind:value={service}>
			<option value="github">GitHub</option>
			<option value="gitlab">GitLab</option>
		</select>
		<label>
			Group Path:
			<input type="text" bind:value={group_path} required />
		</label>
		<label>
			Repository URL:
			<input type="text" bind:value={repoUrl} required />
		</label>
		<label>
			Username:
			<input type="text" bind:value={username} required />
		</label>
		<label>
			Password:
			<input type="password" bind:value={password} required />
		</label>
		<label>
			Token:
			<input type="password" bind:value={token} required />
		</label>
		<button type="submit">Clone Repository</button>
	</form>
	<ul>
		{#if repos.length > 0}
			{#each repos as repo}
				<li><a href={repo.url} target="_blank">{repo.name}</a></li>
			{/each}
		{/if}
	</ul>
</main>

<style>
	main {
		padding: 2rem;
		max-width: 600px;
		margin: 0 auto;
	}

	h1 {
		color: #ff3e00;
		text-transform: uppercase;
		font-size: 4em;
		font-weight: 100;
	}

	label {
		display: block;
		margin-bottom: 1rem;
	}

	input {
		width: 100%;
		padding: 0.5rem;
		margin-top: 0.5rem;
	}

	button {
		padding: 0.5rem 1rem;
	}

	ul {
		list-style-type: none;
		padding: 0;
	}

	li {
		margin: 0.5rem 0;
	}

	@media (min-width: 640px) {
		main {
			max-width: none;
		}
	}
</style>
